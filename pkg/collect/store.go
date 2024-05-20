package collect

import (
	"context"
	"encoding/json"
	"errors"

	"template/pkg/storage"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

//go:generate mockgen -source=./store.go -destination=./store_mock.go -package=collect

type MetricWithUpdate struct {
	Metric    dto.Metric `json:"metric"`
	Name      string     `json:"name"`
	Help      string     `json:"help"`
	UpdatedAt int64      `json:"updated_at"`
}

var ErrNotFound = errors.New("not found")

type Store interface {
	Get(ctx context.Context, id uint64, out *MetricWithUpdate) error
	Put(ctx context.Context, id uint64, input *MetricWithUpdate) error
	List(ctx context.Context, opts *Opts) ([]*MetricWithUpdate, error)
}

type Opts struct {
	Name         string
	Help         string
	UpdatedIndex int64
	Label        prometheus.Labels
}

func NewStore(db *storage.DB) *store {
	return &store{
		DB: db,
	}
}

type store struct {
	*storage.DB
}

type metric struct {
	storage.Base
	Name         string         `json:"name"`
	Help         string         `json:"help"`
	UpdatedIndex int64          `json:"updated_index"`
	Metric       datatypes.JSON `json:"metric"`
}

func (my *store) Get(ctx context.Context, id uint64, out *MetricWithUpdate) error {
	var data metric
	if err := my.WithContext(ctx).
		Where("id = ?", id).
		First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := json.Unmarshal(data.Metric, &out.Metric); err != nil {
		return err
	}
	out.Name = data.Name
	out.Help = data.Help
	out.UpdatedAt = data.UpdatedIndex
	return nil
}

func (my *store) Put(ctx context.Context, id uint64, input *MetricWithUpdate) error {
	jsonBytes, err := json.Marshal(input.Metric)
	if err != nil {
		return err
	}
	return my.WithContext(ctx).FirstOrCreate(&metric{
		Base: storage.Base{
			SnowID: storage.SnowID{
				ID: id,
			},
		},
		Name:         input.Name,
		Help:         input.Help,
		UpdatedIndex: input.UpdatedAt,
		Metric:       jsonBytes,
	}).Error
}

func (my *store) List(ctx context.Context, opts *Opts) ([]*MetricWithUpdate, error) {
	query := my.WithContext(ctx).Model(&metric{})
	if opts != nil {
		if opts.Name != "" {
			query = query.Where("name = ?", opts.Name)
		}
		if opts.Help != "" {
			query = query.Where("help = ?", opts.Help)
		}
		if opts.UpdatedIndex != 0 {
			query = query.Where("updated_index > ?", opts.UpdatedIndex)
		}
	}
	var out []*metric
	if err := query.Find(&out).Error; err != nil {
		return nil, err
	}
	list := make([]*MetricWithUpdate, 0, len(out))
	for _, v := range out {
		out := MetricWithUpdate{
			Metric:    dto.Metric{},
			Name:      v.Name,
			Help:      v.Help,
			UpdatedAt: v.UpdatedIndex,
		}
		err := json.Unmarshal([]byte(v.Metric), &out.Metric)
		if err != nil {
			return nil, err
		}
		if opts != nil {
			selectFlag := true
			for k, l := range opts.Label {
				for _, o := range out.Metric.Label {
					if k == *o.Name && l != *o.Value {
						selectFlag = false
						break
					}
				}
			}
			if !selectFlag {
				continue
			}
		}
		list = append(list, &out)
	}
	return list, nil
}
