package repository

type Storage interface {
	Get(metricType, metricName string) (string, bool)
	Update(metricType, metricName, value string)
}
