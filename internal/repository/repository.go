package repository

type Storage interface {
	Create(metricType, metricName, value string) error
	Get(metricType, metricName string) (string, error)
	GetAll() ([][]string, error)
	Update(metricType, metricName, value string) error
	Delete(metricType, metricName string) error
}
