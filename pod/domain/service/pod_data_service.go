package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"pod/domain/model"
	"pod/domain/repository"
	"pod/proto/pod"

	"git.imooc.com/hedonwang/commom"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IPodDataService interface {
	AddPod(newPod *model.Pod) (int64, error)
	DeletePod(podID int64) error
	UpdatePod(updatedPod *model.Pod) error
	FindPodByID(podID int64) (*model.Pod, error)
	FindAllPod() ([]model.Pod, error)
	CreatePodToK8s(podInfo *pod.PodInfo) error
	DeletePodFromK8s(pod *model.Pod) error
	UpdatePodToK8s(podInfo *pod.PodInfo) error
}

type PodDataService struct {
	PodRepository repository.PodRepository
	K8sClientSet  *kubernetes.Clientset
	Deployment    *v1.Deployment
}

// NewPodDataService 创建一个新的 PodDataService
func NewPodDataService(podRepository repository.PodRepository, clientSet *kubernetes.Clientset) IPodDataService {
	return &PodDataService{
		PodRepository: podRepository,
		K8sClientSet:  clientSet,
		Deployment:    &v1.Deployment{},
	}
}

func (ps *PodDataService) AddPod(newPod *model.Pod) (int64, error) {
	return ps.PodRepository.CreatePod(newPod)
}

func (ps *PodDataService) DeletePod(podID int64) error {
	return ps.PodRepository.DeletePodByID(podID)
}

func (ps *PodDataService) UpdatePod(updatedPod *model.Pod) error {
	return ps.PodRepository.UpdatePod(updatedPod)
}

func (ps *PodDataService) FindPodByID(podID int64) (*model.Pod, error) {
	return ps.PodRepository.FindPodByID(podID)
}

func (ps *PodDataService) FindAllPod() ([]model.Pod, error) {
	return ps.PodRepository.FindAll()
}

// CreatePodToK8s 同步创建 Pod 到 k8s
func (ps *PodDataService) CreatePodToK8s(podInfo *pod.PodInfo) error {

	ctx := context.Background()

	// 初始化 Deployment
	ps.setDeployment(podInfo)

	// 创建 Deployment
	_, err := ps.K8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).Get(ctx, podInfo.PodName, v12.GetOptions{})
	if err != nil {
		_, err = ps.K8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).Create(ctx, ps.Deployment, v12.CreateOptions{})
		if err != nil {
			commom.Error(err)
			return err
		}
		commom.Info("create Deployment successfully")
	} else {
		commom.Error("Pod [" + podInfo.PodName + "] existed")
		return errors.New("Pod [" + podInfo.PodName + "] existed")
	}

	return nil
}

// DeletePodFromK8s 同步删除 k8s 中的 Pod
func (ps *PodDataService) DeletePodFromK8s(pod *model.Pod) error {

	ctx := context.Background()

	_, err := ps.K8sClientSet.AppsV1().Deployments(pod.PodNamespace).Get(ctx, pod.PodName, v12.GetOptions{})
	if err != nil {
		commom.Error(err)
		return errors.New("pod [" + pod.PodName + "] not exists")
	}

	err = ps.K8sClientSet.AppsV1().Deployments(pod.PodNamespace).Delete(ctx, pod.PodName, v12.DeleteOptions{})
	if err != nil {
		commom.Error(err)
		return err
	}

	err = ps.DeletePod(pod.ID)
	if err != nil {
		commom.Error(err)
		return err
	}

	commom.Info("delete pod [" + pod.PodName + "] successfully")

	return nil
}

// UpdatePodToK8s 同步更新 k8s 中的 Pod
func (ps *PodDataService) UpdatePodToK8s(podInfo *pod.PodInfo) error {

	ctx := context.Background()

	ps.setDeployment(podInfo)

	_, err := ps.K8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).Get(ctx, podInfo.PodName, v12.GetOptions{})
	if err != nil {
		commom.Error(err)
		return errors.New("pod [" + podInfo.PodName + "] not exists")
	}

	_, err = ps.K8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).Update(ctx, ps.Deployment, v12.UpdateOptions{})
	if err != nil {
		commom.Error(err)
		return err
	}

	commom.Info("update pod [" + podInfo.PodName + "] successfully")

	return nil
}

// SetDeployment 通过 podInfo 来包装 Deployment 的元数据
func (ps *PodDataService) setDeployment(podInfo *pod.PodInfo) {
	ps.Deployment = &v1.Deployment{
		TypeMeta: v12.TypeMeta{
			Kind:       "deployment",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:      podInfo.PodName,
			Namespace: podInfo.PodNamespace,
			Labels: map[string]string{
				"app-name": podInfo.PodName,
				"author":   "hedon",
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &podInfo.PodReplicas, // 副本个数
			Selector: &v12.LabelSelector{
				MatchLabels: map[string]string{
					"app-name": podInfo.PodName,
				},
				MatchExpressions: nil,
			},
			Template: v13.PodTemplateSpec{
				ObjectMeta: v12.ObjectMeta{
					Labels: map[string]string{
						"app-name": podInfo.PodName,
					},
				},
				Spec: v13.PodSpec{
					Containers: []v13.Container{
						{
							Name:            podInfo.PodName,
							Image:           podInfo.PodImage,
							Ports:           ps.getContainerPorts(podInfo),
							Env:             ps.getContainerEnvs(podInfo),
							Resources:       ps.getContainerResources(podInfo),
							ImagePullPolicy: ps.getContainerImagePullPolicy(podInfo),
						},
					},
				},
			},
		},
	}
	ps.Deployment.Name = podInfo.PodName
}

// getContainerPorts 包装容器端口
func (ps *PodDataService) getContainerPorts(podInfo *pod.PodInfo) []v13.ContainerPort {
	res := make([]v13.ContainerPort, 0, len(podInfo.PodPorts))
	for _, v := range podInfo.PodPorts {
		res = append(res, v13.ContainerPort{
			Name:          "port-" + strconv.Itoa(int(v.ContainerPort)),
			ContainerPort: v.ContainerPort,
			Protocol:      ps.getContainerProtocol(v.Protocol),
		})
	}
	return res
}

// getContainerProtocol 包装容器通讯协议
func (ps *PodDataService) getContainerProtocol(protocol string) v13.Protocol {
	switch strings.ToLower(protocol) {
	case "tcp":
		return "TCP"
	case "udp":
		return "UDP"
	case "sctp":
		return "SCTP"
	default:
		return "TCP"
	}
}

// getContainerEnvs 包装容器环境变量
func (ps *PodDataService) getContainerEnvs(podInfo *pod.PodInfo) []v13.EnvVar {
	res := make([]v13.EnvVar, 0, len(podInfo.PodEnvs))
	for _, v := range podInfo.PodEnvs {
		res = append(res, v13.EnvVar{
			Name:  v.EnvKey,
			Value: v.EnvValue,
		})
	}
	return res
}

// getContainerResources 包装容器资源
func (ps *PodDataService) getContainerResources(podInfo *pod.PodInfo) v13.ResourceRequirements {
	source := v13.ResourceRequirements{}

	// 最大资源使用限制
	source.Limits = v13.ResourceList{
		"cpu":    resource.MustParse(strconv.FormatFloat(float64(podInfo.PodCpuMax), 'f', 6, 64)),
		"memory": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodMemoryMax), 'f', 6, 64)),
	}

	// 最小资源使用限制
	// TODO
	source.Requests = v13.ResourceList{}

	return source
}

// getContainerImagePullPolicy 包装容器镜像拉取策略
func (ps *PodDataService) getContainerImagePullPolicy(podInfo *pod.PodInfo) v13.PullPolicy {
	switch strings.ToLower(podInfo.PodPullPolicy) {
	case "always":
		return "Always"
	case "never":
		return "never"
	case "ifnotpresent":
		return "IfNotPresent"
	default:
		return "Always"
	}
}
