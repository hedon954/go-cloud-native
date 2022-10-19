package model

type Pod struct {
	ID           int64  `json:"id" gorm:"primaryKey;notNull;autoIncrement"`
	PodName      string `json:"pod_name" gorm:"uniqueIndex;notNull"`
	PodNamespace string `json:"pod_namespace"`

	// Pod 所属团队
	PodTeamID int64 `json:"pod_team_id"`

	// Pod CPU 使用范围
	PodCpuMin float32 `json:"pod_cpu_min"`
	PodCpuMax float32 `json:"pod_cpu_max"`

	// Pod 副本数
	PodReplicas int32 `json:"pod_replicas"`

	// Pod 内存使用范围
	PodMemoryMin float32 `json:"pod_memory_min"`
	PodMemoryMax float32 `json:"pod_memory_max"`

	// Pod 开发的端口
	PodPorts []PodPort `json:"pod_ports" gorm:"foreignKey:PodID"`

	// Pod 环境变量
	PodEnvs []PodEnv `json:"pod_envs" gorm:"foreignKey:PodID"`

	// 镜像拉取策略
	PodPullPolicy string `json:"pod_pull_policy"`

	// 重启策略
	PodRestartPolicy string `json:"pod_restart_policy"`

	// 发布策略
	PodReleasePolicy string `json:"pod_release_policy"`

	// 镜像名称 + tag
	PodImage string `json:"pod_image"`

	// TODO：挂盘、域名设置、etc...
}
