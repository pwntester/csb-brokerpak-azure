package helpers

import (
	"encoding/json"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type ServiceInstance struct {
	offering string
	name     string
}

func CreateService(offering, plan string, parameters ...interface{}) ServiceInstance {
	name := RandomName(offering, plan)
	args := append([]string{"create-service", offering, plan, name}, serviceParameters(parameters)...)

	session := StartCF(args...)
	Eventually(session, time.Minute).Should(Exit(0))

	Eventually(func() string {
		out, _ := CF("service", name)
		Expect(out).NotTo(MatchRegexp(`status:\s+create failed`))
		return out
	}, 30*time.Minute, 30*time.Second).Should(MatchRegexp(`status:\s+create succeeded`))

	return ServiceInstance{
		name:     name,
		offering: offering,
	}
}

func (s ServiceInstance) Delete() {
	CF("delete-service", "-f", s.name)
	Eventually(func() string {
		out, _ := CF("services")
		return out
	}, 30*time.Minute, 30*time.Second).ShouldNot(ContainSubstring(s.name))
}

func (s ServiceInstance) Bind(app AppInstance) Binding {
	name := RandomName()
	CF("bind-service", app.name, s.name, "--binding-name", name)

	return Binding{
		serviceInstance: s,
		bindingName:     name,
		appInstance:     app,
	}
}

func (s ServiceInstance) CreateKey() ServiceKey {
	name := RandomName()
	CF("create-service-key", s.name, name)

	return ServiceKey{
		name:            name,
		serviceInstance: s,
	}
}

func (s ServiceInstance) Name() string {
	return s.name
}

func serviceParameters(parameters []interface{}) []string {
	if len(parameters) > 0 {
		switch p := parameters[0].(type) {
		case string:
			return []string{"-c", p}
		default:
			params, err := json.Marshal(p)
			Expect(err).NotTo(HaveOccurred())
			return []string{"-c", string(params)}
		}
	}

	return []string{}
}