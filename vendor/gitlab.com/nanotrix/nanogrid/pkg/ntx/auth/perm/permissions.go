package perm

import (
	"path"
	"strings"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type (
	Permissions interface {
		UseRole(role string) bool
		Role() string
		RunTask(id string) bool
		UseService(id string) bool
		CreateTask() bool
		SystemAdmin() bool
		SystemMonitor() bool
		Encode() []string
	}

	permissions struct {
		role          string
		tasks         []string // task patterns
		services      []string // task patterns
		createTask    bool
		systemAdmin   bool
		systemMonitor bool
		allTasks      bool
		allServices   bool
	}
)

const (
	unlimitedRole = "unlimited"

	permissionCreateCustomTask = "task:create"
	permissionSystemAdmin      = "system:admin"
	permissionSystemMonitor    = "system:monitor"
	permissionUseServicePrefix = "service:"
	permissionTaskRunPrefix    = "task:run:"
	permissionAllowAll         = "*"
	permissionUseRolePrefix    = "role:"
)

func RunTask(idOrPath string) string {
	return permissionTaskRunPrefix + idOrPath
}

func RunAnyTask() string {
	return permissionTaskRunPrefix + permissionAllowAll
}

func CustomTask() string {
	return permissionCreateCustomTask
}

func SystemAdmin() string {
	return permissionSystemAdmin
}

func SystemMonitor() string {
	return permissionSystemMonitor
}

func UseRole(role string) string {
	return permissionUseRolePrefix + role
}

func UseService(name string) string {
	return permissionUseServicePrefix + name
}
func UseAnyService() string {
	return permissionUseServicePrefix + permissionAllowAll
}

func UnlimitedRole() string {
	return unlimitedRole
}

func Parse(perms ...string) Permissions {
	resp := &permissions{role: unlimitedRole}

	perms = util.DeduplicateStringSlice(perms)

	for _, p := range perms {
		if p == permissionCreateCustomTask {
			resp.createTask = true
		} else if p == permissionSystemAdmin {
			resp.systemAdmin = true
		} else if p == permissionSystemMonitor {
			resp.systemMonitor = true
		} else if strings.HasPrefix(p, permissionUseRolePrefix) {
			role := strings.TrimPrefix(p, permissionUseRolePrefix)
			if role != "" {
				resp.role = role
			}
		} else if strings.HasPrefix(p, permissionTaskRunPrefix) {
			tsk := strings.TrimPrefix(p, permissionTaskRunPrefix)
			if tsk == permissionAllowAll {
				resp.allTasks = true
			} else {
				resp.tasks = append(resp.tasks, tsk)
			}
		} else if strings.HasPrefix(p, permissionUseServicePrefix) {
			svc := strings.TrimPrefix(p, permissionUseServicePrefix)
			if svc == permissionAllowAll {
				resp.allServices = true
			} else {
				resp.services = append(resp.services, svc)
			}
		}
	}

	return resp
}

func (p *permissions) Encode() []string {
	var out []string

	if p.CreateTask() {
		out = append(out, CustomTask())
	}

	if p.SystemAdmin() {
		out = append(out, SystemAdmin())
	} else if p.SystemMonitor() { // admin has always monitor role
		out = append(out, SystemMonitor())
	}

	if p.allTasks {
		out = append(out, RunAnyTask())
	} else {
		for _, t := range p.tasks {
			out = append(out, RunTask(t))
		}
	}

	if p.allServices {
		out = append(out, UseAnyService())
	} else {
		for _, s := range p.services {
			out = append(out, UseService(s))
		}
	}

	if p.role != unlimitedRole {
		out = append(out, UseRole(p.role))
	}

	return out
}

func (p *permissions) Role() string {
	return p.role
}

func (p *permissions) UseRole(role string) bool {
	return role == p.role
}

func (p *permissions) SystemAdmin() bool {
	return p.systemAdmin
}

func (p *permissions) SystemMonitor() bool {
	return p.systemAdmin || p.systemMonitor
}

func (p *permissions) CreateTask() bool {
	return p.createTask
}

func (p *permissions) RunTask(id string) bool {
	if p.allTasks {
		return true
	}

	for _, tp := range p.tasks {
		if ok, _ := path.Match(tp, id); ok {
			return true
		}
	}
	return false
}

func (p *permissions) UseService(id string) bool {
	if p.allServices {
		return true
	}

	for _, s := range p.services {
		if ok, _ := path.Match(s, id); ok {
			return true
		}
	}
	return false
}

func Sort(perms ...string) []string {
	return Parse(perms...).Encode()
}
