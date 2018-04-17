package initializer

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	components "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	componentsclientset "github.com/goblain/mariadb-operator/pkg/generated/clientset/versioned"
	"github.com/goblain/mariadb-operator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	defaultKubeAPIRequestTimeout = 30 * time.Second
)

type Initializer struct {
	Hostname     string
	clientConfig *rest.Config
	// client           *kubernetes.Clientset
	componentsClient *componentsclientset.Clientset
	// apiExtensionsClient *apiextensionsclientset.Clientset
	logger    *logrus.Entry
	name      string
	namespace string
}

func (i *Initializer) Run() {

	// Take care of termination by signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		logrus.Infof("received signal: %v, exiting", <-c)
		os.Exit(1)
	}()

	var err error

	i.name = os.Getenv("MARIADBCLUSTER_NAME")
	i.namespace = os.Getenv("MARIADBCLUSTER_NAMESPACE")
	if i.Hostname, err = os.Hostname(); err != nil {
		panic(err.Error())
	}

	logrus.SetLevel(logrus.DebugLevel)
	i.logger = logrus.WithField("namespace", i.namespace).WithField("name", i.name)
	i.logger.Debug("Debug logging enabled")

	i.clientConfig, err = InClusterConfig()
	if err != nil {
		panic(err)
	}
	i.clientConfig.Timeout = defaultKubeAPIRequestTimeout
	i.componentsClient = componentsclientset.NewForConfigOrDie(i.clientConfig)

	mdbc := i.getMariaDBCluster()

	writeConfig(mdbc)

	hostname, _ := os.Hostname()

	if mdbc.Status.Phase == components.PhaseRecovery {
		// Hold on waiting for recovery stuff to happen
		for true {
			i.logger.Debug("Recovery phase detected, reporting my condition to MariaDBCluster object")
			i.reportToMariaDBCluster(mdbc)
			time.Sleep(time.Second * 5)
			mdbc = i.getMariaDBCluster()
			if mdbc.Status.Stage == "PrimaryRecovered" {
				// Primary recovered, release from the stasis for cluster rejoin
				mdbc.Status.StatefulSetPodConditions = nil
				break
			} else if hostname == mdbc.Status.BootstrapFrom {
				// Marked for primary recovery, release and bootstrap new cluster
				setSafeToBootstrap()
				break
			}
		}
	}

	writeConfig(mdbc)
}

func setSafeToBootstrap() {
	state := []byte(getStateString())
	re := regexp.MustCompile(`(safe_to_bootstrap:.*)(0)`)
	newState := re.ReplaceAll(state, []byte(`$1 1`))
	ioutil.WriteFile("/var/lib/mysql/grastate.dat", newState, 0440)
}

func writeConfig(mdbc *components.MariaDBCluster) {
	var mdbConfig *components.MariaDBConfig
	hostname, _ := os.Hostname()
	if hostname == mdbc.Status.BootstrapFrom {
		mdbConfig = &components.MariaDBConfig{
			Name:                 mdbc.GetServerName(),
			WSREPEndpoints:       nil,
			WSREPProviderOptions: "pc.bootstrap=true",
		}
	} else {
		mdbConfig = &components.MariaDBConfig{
			Name:                 mdbc.GetServerName(),
			WSREPEndpoints:       mdbc.GetWSREPEndpoints(),
			WSREPProviderOptions: "",
		}
	}

	operatorCnf, err := mdbConfig.Render()
	if err != nil {
		panic("Can't render template : " + err.Error())
	}

	err = ioutil.WriteFile("/etc/mysql/conf.d/operator.cnf", []byte(operatorCnf), 0444)
	if err != nil {
		panic(err.Error())
	}
	config := `[mysqld]` + "\n" + mdbc.Spec.ServerConfig
	logrus.Debug(config)
	err = ioutil.WriteFile("/etc/mysql/conf.d/user.cnf", []byte(config), 0444)
	if err != nil {
		panic(err.Error())
	}
}

func (i *Initializer) getMariaDBCluster() *components.MariaDBCluster {
	mdbc, err := i.componentsClient.Components().MariaDBClusters(i.namespace).Get(i.name, metav1.GetOptions{})
	if err != nil {
		i.logger.Error("where is my cluster?!? : " + err.Error())
		panic("where is my cluster?!? : " + err.Error())
	}
	return mdbc
}

func (i *Initializer) reportToMariaDBCluster(mdbc *components.MariaDBCluster) {
	current := i.getMariaDBCluster()
	expected := current.DeepCopy()

	version, uuid, seqno, safeToBootstrap := parseGRAState()
	if seqno <= 0 {
		// got some stupid SeqNo, proceed into recovery
		uuidRec, seqnoRec, err := recoverGRAStateUuidSeqNo()
		if err == nil && uuid == uuidRec {
			seqno = seqnoRec
		} else {
			panic("UUID missmatch")
		}
	}

	podCondition := components.PodCondition{
		Hostname: i.Hostname,
		Reported: metav1.Now(),
		GRAState: components.PodConditionGRAState{Version: version, UUID: uuid, SeqNo: seqno, SafeToBootstrap: safeToBootstrap},
	}
	match := false
	for k, v := range expected.Status.StatefulSetPodConditions {
		if v.Hostname == podCondition.Hostname {
			match = true
			if !reflect.DeepEqual(v.GRAState, podCondition.GRAState) {
				logrus.Debug("Entry found and status needs update")
				expected.Status.StatefulSetPodConditions[k].Hostname = podCondition.Hostname
				expected.Status.StatefulSetPodConditions[k].Reported = podCondition.Reported
				expected.Status.StatefulSetPodConditions[k].GRAState = podCondition.GRAState
			} else {
				logrus.Debug("Entry found and no status update required")
			}
			break
		}
	}
	if !match {
		expected.Status.StatefulSetPodConditions = append(expected.Status.StatefulSetPodConditions, podCondition)
	}
	util.CheckAndPatchMariaDBCluster(current, expected, i.componentsClient.Components(), i.logger)
}

func recoverGRAStateUuidSeqNo() (string, int64, error) {
	logrus.Debug("Recovering wsrep state")
	cmd := exec.Command("su", "mysql", "-c", "/usr/sbin/mysqld --wsrep-recover")
	out, _ := cmd.CombinedOutput()
	re := regexp.MustCompile(`WSREP: Recovered position:\s*([0-9a-z-]*):(\d+)`)
	result := re.FindStringSubmatch(string(out))
	if len(result) > 1 {
		seqno, _ := strconv.ParseInt(result[2], 10, 64)
		return result[1], seqno, nil
	}
	return "", int64(0), fmt.Errorf("failed to recover")
}

func getStateString() string {
	stateString, err := ioutil.ReadFile("/var/lib/mysql/grastate.dat")
	if err != nil {
		panic("missing grastate.dat : " + err.Error())
	}
	return string(stateString)
}

func parseGRAState() (string, string, int64, int) {
	stateString := getStateString()
	var safeToBootstrap int
	var seqno int64
	var uuid, version string
	var re *regexp.Regexp
	var result []string

	logrus.Debug("stateString : " + stateString)

	re = regexp.MustCompile(`version:\s*([0-9\.]*)`)
	result = re.FindStringSubmatch(stateString)
	if len(result) > 1 {
		version = result[1]
	} else {
		panic("Version missing")
	}

	logrus.Debug("version " + version)

	re = regexp.MustCompile(`uuid:\s*([A-Za-z0-9-]*)`)
	result = re.FindStringSubmatch(stateString)
	if len(result) > 1 {
		uuid = result[1]
	} else {
		panic("UUID missing")
	}
	logrus.Debug("uuid " + uuid)

	re = regexp.MustCompile(`seqno:\s*([-]?\d+)`)
	result = re.FindStringSubmatch(stateString)
	if len(result) > 1 {
		seqno, _ = strconv.ParseInt(result[1], 10, 64)
	} else {
		panic("SeqNo missing")
	}
	logrus.Debug("version " + string(seqno))

	re = regexp.MustCompile(`safe_to_bootstrap:\s*(\d)`)
	result = re.FindStringSubmatch(string(stateString))
	if len(result) > 1 {
		safeToBootstrap, _ = strconv.Atoi(result[1])
	} else {
		logrus.Warn("safe_to_bootstrap missing")
	}
	logrus.Debug("safetoBootstrap " + string(safeToBootstrap))

	return version, uuid, seqno, safeToBootstrap
}

func InClusterConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
