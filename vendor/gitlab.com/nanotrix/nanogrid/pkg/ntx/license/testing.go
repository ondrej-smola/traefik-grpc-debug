package license

import (
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth/perm"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

func TestClient() *Client {
	c := &Client{
		ClusterId:   util.NewUUID(),
		CustomerId:  "customer1",
		Label:       "label",
		Permissions: []string{perm.RunAnyTask()},
		Tasks:       "http://localhost:8081/tasks",
		Limits: &Limits{
			MonthlyCredit: 1000,
		},
		Capabilities:       Capabilities{Client_Capabilities_CUSTOM_TASKS},
		Requirements:       Requirements{Client_Requirements_MESOS_HOSTNAME_CHECK},
		MaxLicenseDuration: 3600,
		Expire:             timeutil.NowUTCUnixSeconds() + 7200,
	}

	if err := c.Valid(); err != nil {
		panic(errors.Wrap(err, "test client"))
	}

	return c
}

func TestTask() *Task {
	t := &Task{
		Service: "task-" + util.RandString(16),
		Tags:    []string{"tag1", "tag2"},
		Profiles: []*Task_Profile{
			{
				Id:     "test-task-" + util.RandString(16),
				Labels: []string{"label-" + util.RandString(8)},
				Task:   scheduler.TestTaskConfiguration(),
			},
		},
	}

	if err := t.Valid(); err != nil {
		panic(errors.Wrap(err, "test task"))
	}

	return t
}

func TestCluster() *Cluster {
	c := &Cluster{
		Id:       util.NewUUID(),
		Created:  timeutil.NowUTCUnixSeconds(),
		Key:      util.RandString(16),
		Hostname: "localhost",
	}

	if err := c.Valid(); err != nil {
		panic(errors.Wrap(err, "test cluster"))
	}

	return c
}
