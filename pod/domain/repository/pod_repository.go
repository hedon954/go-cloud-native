package repository

import (
	"pod/domain/model"

	"github.com/jinzhu/gorm"
)

type IPodRepository interface {
	InitTable() error
	FindPodByID(podID int64) (*model.Pod, error)
	CreatePod(pod *model.Pod) (int64, error)
	DeletePodByID(podID int64) error
	UpdatePod(pod *model.Pod) error
	FindAll() ([]model.Pod, error)
}

// NewPodRepository 创建一个 PodRepository
func NewPodRepository(db *gorm.DB) IPodRepository {
	return &PodRepository{
		mysqlDB: db,
	}
}

type PodRepository struct {
	mysqlDB *gorm.DB
}

func (p *PodRepository) InitTable() error {
	return p.mysqlDB.CreateTable(&model.Pod{}, &model.PodPort{}, &model.PodEnv{}).Error
}

func (p *PodRepository) FindPodByID(podID int64) (*model.Pod, error) {
	pod := &model.Pod{}
	return pod, p.mysqlDB.Preload("PodEnv").Preload("PodPort").First(pod, podID).Error
}

func (p *PodRepository) CreatePod(pod *model.Pod) (int64, error) {
	return pod.ID, p.mysqlDB.Create(pod).Error
}

func (p *PodRepository) DeletePodByID(podID int64) error {
	tx := p.mysqlDB.Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	err := p.mysqlDB.Where("id = ?", podID).Delete(&model.Pod{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = p.mysqlDB.Where("pod_id = ?", podID).Delete(&model.PodEnv{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = p.mysqlDB.Where("pod_id = ?", podID).Delete(&model.PodPort{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (p *PodRepository) UpdatePod(pod *model.Pod) error {
	return p.mysqlDB.Model(pod).Update(pod).Error
}

func (p *PodRepository) FindAll() ([]model.Pod, error) {
	res := make([]model.Pod, 0)
	return res, p.mysqlDB.Find(&res).Error
}
