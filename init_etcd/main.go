package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/mshakery/ServerlessController/etcdMiddleware"
	"github.com/mshakery/ServerlessController/protos"
)

func main() {
	addUsers, addRoles, addRoleBinding, addNode := true, true, true, true

	cli, err := etcdMiddleware.ConnectToEtcd()
	if err != nil {
		panic(err)
	}

	if addUsers {
		users := []protos.User{
			{Token: "xaxaxaxa", Uid: "0", UserName: "masoud"},
			{Token: "ihihihih", Uid: "1", UserName: "roy"},
			{Token: "lolololo", Uid: "2", UserName: "arshia"},
		}

		for _, user := range users {
			ku := fmt.Sprintf("/cluster/resources/user/%s", user.Token)
			err = etcdMiddleware.WriteToEtcdFromPb(cli, context.Background(), ku, &user)
			if err != nil {
				fmt.Printf("ik kan schrijf het %s kunde niet! wat verveland! want: %s", user.Uid, err.Error())
				return
			}
			fmt.Printf("ik schrijf het %s kunde op de database. Wat leuk! ", user.UserName)
		}
	}

	if addRoles {
		roles := []protos.Role{
			{
				Metadata: &protos.Metadata{
					Name:      "role1",
					Namespace: "default",
					Uid:       "0",
				},
				Rules: []*protos.PolicyRule{
					{
						Verbs: []string{"create", "retrieve", "update", "delete"},
						Resources: []string{
							"pod",
							"deployment",
							"node",
							"role",
							"user",
							"role_binding",
							"hpa",
						},
					},
				},
			},
			{
				Metadata: &protos.Metadata{
					Name:      "role2",
					Namespace: "default",
					Uid:       "1",
				},
				Rules: []*protos.PolicyRule{
					{
						Verbs: []string{"create"},
						Resources: []string{
							"user",
						},
					},
				},
			},
		}

		for _, role := range roles {
			ku := fmt.Sprintf("/cluster/resources/role/%s", role.Metadata.Name)
			err = etcdMiddleware.WriteToEtcdFromPb(cli, context.Background(), ku, &role)
			if err != nil {
				fmt.Printf("fail. %s", err.Error())
				return
			}
			fmt.Printf("soccess ")
		}
	}

	if addRoleBinding {
		roleBindings := []protos.RoleBinding{
			{
				Metadata: &protos.Metadata{
					Name:      "roleBinding1",
					Namespace: "default",
					Uid:       "0",
				},
				RoleRef:  "role1",
				Subjects: []string{"0", "2"},
			},
			{
				Metadata: &protos.Metadata{
					Name:      "roleBinding2",
					Namespace: "default",
					Uid:       "1",
				},
				RoleRef:  "role2",
				Subjects: []string{"1"},
			},
		}

		for _, roleBinding := range roleBindings {
			ku := fmt.Sprintf("/cluster/resources/role_binding/%s", roleBinding.Metadata.Name)
			err = etcdMiddleware.WriteToEtcdFromPb(cli, context.Background(), ku, &roleBinding)
			if err != nil {
				fmt.Printf("fail. %s", err.Error())
				return
			}
			fmt.Printf("soccess ")
		}
	}

	if addNode {
		podlist := protos.PodList{
			PodNames: []string{},
		}
		nodes := []protos.Node{
			{
				Metadata: &protos.Metadata{
					Name:      "kubelet-0",
					Namespace: "default",
					Uid:       "kubelet-0",
				},
				Spec: &protos.NodeSpec{
					Unschedulable: &protos.Unschedulable{
						Condition: false,
					},
				},
				Status: &protos.NodeStatus{
					Condition: &protos.Condition{
						LastUpdate: &timestamp.Timestamp{Seconds: 1705062732},
					},
					Capacity: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Allocatable: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Pods: &podlist,
				},
			},
			{
				Metadata: &protos.Metadata{
					Name:      "kubelet-1",
					Namespace: "default",
					Uid:       "kubelet-1",
				},
				Spec: &protos.NodeSpec{
					Unschedulable: &protos.Unschedulable{
						Condition: false,
					},
				},
				Status: &protos.NodeStatus{
					Condition: &protos.Condition{
						LastUpdate: &timestamp.Timestamp{Seconds: 1705062732},
					},
					Capacity: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Allocatable: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Pods: &podlist,
				},
			},
			{
				Metadata: &protos.Metadata{
					Name:      "kubelet-2",
					Namespace: "default",
					Uid:       "kubelet-2",
				},
				Spec: &protos.NodeSpec{
					Unschedulable: &protos.Unschedulable{
						Condition: false,
					},
				},
				Status: &protos.NodeStatus{
					Condition: &protos.Condition{
						LastUpdate: &timestamp.Timestamp{Seconds: 1705062732},
					},
					Capacity: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Allocatable: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Pods: &podlist,
				},
			},
			{
				Metadata: &protos.Metadata{
					Name:      "kubelet-3",
					Namespace: "default",
					Uid:       "kubelet-3",
				},
				Spec: &protos.NodeSpec{
					Unschedulable: &protos.Unschedulable{
						Condition: false,
					},
				},
				Status: &protos.NodeStatus{
					Condition: &protos.Condition{
						LastUpdate: &timestamp.Timestamp{Seconds: 1705062732},
					},
					Capacity: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Allocatable: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Pods: &podlist,
				},
			},
			{
				Metadata: &protos.Metadata{
					Name:      "kubelet-4",
					Namespace: "default",
					Uid:       "kubelet-4",
				},
				Spec: &protos.NodeSpec{
					Unschedulable: &protos.Unschedulable{
						Condition: false,
					},
				},
				Status: &protos.NodeStatus{
					Condition: &protos.Condition{
						LastUpdate: &timestamp.Timestamp{Seconds: 1705062732},
					},
					Capacity: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Allocatable: &protos.Capacity{
						Resources: map[string]string{
							"cpu":    "10000",
							"memory": "8192000",
						},
					},
					Pods: &podlist,
				},
			},
		}

		for _, node := range nodes {
			ku := fmt.Sprintf("/cluster/resources/node/%s", node.Metadata.Name)
			err = etcdMiddleware.WriteToEtcdFromPb(cli, context.Background(), ku, &node)
			if err != nil {
				fmt.Printf("fail. %s", err.Error())
				return
			}
			ku2 := fmt.Sprintf("/cluster/resources/node/%s/allocatable", node.Metadata.Name)
			err = etcdMiddleware.WriteToEtcdFromPb(cli, context.Background(), ku2, node.GetStatus().GetCapacity())
			if err != nil {
				fmt.Printf("fail. %s", err.Error())
				return
			}
			ku3 := fmt.Sprintf("/cluster/resources/node/%s/unschedulable", node.Metadata.Name)
			err = etcdMiddleware.WriteToEtcdFromPb(cli, context.Background(), ku3, node.Spec.Unschedulable)
			if err != nil {
				fmt.Printf("fail. %s", err.Error())
				return
			}
			fmt.Printf("node soccess ")
		}
	}
}
