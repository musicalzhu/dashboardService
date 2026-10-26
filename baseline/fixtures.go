package baseline

import "strings"

// AllServiceCombinations 返回 POC 使用的服务组合 fixture，包含 AS、TS、SS、DS 的全部 15 种非空组合。
// 多种规则分发方案的演示和测试共同使用该数据，避免重复维护样例。
func AllServiceCombinations() []Server {
	serviceTypes := []ServiceType{AS, TS, SS, DS}
	servers := make([]Server, 0, (1<<len(serviceTypes))-1)
	for mask := 1; mask < 1<<len(serviceTypes); mask++ {
		var services []Service
		var names []string
		for bit, serviceType := range serviceTypes {
			if mask&(1<<(len(serviceTypes)-1-bit)) == 0 {
				continue
			}
			services = append(services, Service{
				ServiceType: serviceType,
				State:       StateNormal,
			})
			names = append(names, string(serviceType))
		}
		servers = append(servers, Server{
			Name:     strings.Join(names, "+"),
			IP:       "1.1.1.1",
			State:    StateNormal,
			Services: services,
		})
	}
	return servers
}
