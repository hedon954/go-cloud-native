package model

// PodEnv Pod 环境变量相关信息
type PodEnv struct {
	ID       int64  `gorm:"primaryKey;notNull;autoIncrement" json:"id"`
	PodID    int64  `json:"pod_id"`
	EnvKey   string `json:"env_key"`
	EnvValue string `json:"env_value"`
}
