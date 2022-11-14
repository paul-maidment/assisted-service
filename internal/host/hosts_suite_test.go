package host_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/assisted-service/internal/common"
)

func TestHost(t *testing.T) {
	RegisterFailHandler(Fail)
	common.InitializeDBTest()
	time.Sleep(10 * time.Second)
	defer common.TerminateDBTest()
	RunSpecs(t, "host state machine tests")
}
