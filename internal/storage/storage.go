package storage

type KeyStorage struct {
	metricType string
	metricName string
}

type MemStorage struct {
	storage map[KeyStorage]string
}

func NewStorage() *MemStorage {
	return &MemStorage{storage: make(map[KeyStorage]string)}
}

func (s *MemStorage) Get(metricType, metricName string) (string, bool) {
	key := KeyStorage{metricType: metricType, metricName: metricName}
	value, ok := s.storage[key]
	return value, ok
}

func (s *MemStorage) Update(metricType, metricName, value string) {
	key := KeyStorage{metricType: metricType, metricName: metricName}
	s.storage[key] = value
}
