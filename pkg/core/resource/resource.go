package resource

import (
	"errors"
	"fmt"
	"sync"
)

// ResourceID представляет уникальный идентификатор ресурса
type ResourceID string

// ResourceType представляет тип ресурса
type ResourceType string

const (
	ResourceTypeTexture ResourceType = "texture"
	ResourceTypeMesh    ResourceType = "mesh"
	ResourceTypeShader  ResourceType = "shader"
	ResourceTypeAudio   ResourceType = "audio"
	ResourceTypeFont    ResourceType = "font"
	ResourceTypeScene   ResourceType = "scene"
	ResourceTypeUnknown ResourceType = "unknown"
)

// Ошибки системы ресурсов
var (
	ErrResourceNotFound   = errors.New("resource not found")
	ErrResourceExists     = errors.New("resource already exists")
	ErrResourceLoading    = errors.New("resource is currently loading")
	ErrInvalidResourceID  = errors.New("invalid resource ID")
	ErrResourceTypeMismatch = errors.New("resource type mismatch")
)

// ResourceState представляет состояние ресурса
type ResourceState int

const (
	ResourceStateUnloaded ResourceState = iota
	ResourceStateLoading
	ResourceStateLoaded
	ResourceStateError
)

// Resource представляет загруженный ресурс
type Resource struct {
	ID           ResourceID
	Path         string
	Type         ResourceType
	Data         interface{}
	State        ResourceState
	RefCount     int
	Size         int64  // Размер в байтах
	Error        error  // Ошибка загрузки, если есть
	mu           sync.RWMutex
}

// AddRef увеличивает счетчик ссылок
func (r *Resource) AddRef() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RefCount++
}

// Release уменьшает счетчик ссылок
func (r *Resource) Release() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RefCount--
	if r.RefCount < 0 {
		r.RefCount = 0
	}
	return r.RefCount
}

// GetRefCount возвращает текущий счетчик ссылок
func (r *Resource) GetRefCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.RefCount
}

// IsLoaded проверяет, загружен ли ресурс
func (r *Resource) IsLoaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.State == ResourceStateLoaded
}

// ResourceLoader интерфейс для загрузчиков ресурсов
type ResourceLoader interface {
	// Load загружает ресурс из файла
	Load(path string) (interface{}, error)

	// Unload выгружает ресурс
	Unload(data interface{}) error

	// GetType возвращает тип ресурсов, которые может загрузить этот загрузчик
	GetType() ResourceType
}

// ResourceManager управляет всеми ресурсами в игре
type ResourceManager struct {
	resources map[ResourceID]*Resource
	loaders   map[ResourceType]ResourceLoader
	cache     map[string]ResourceID // Кеш path -> ResourceID
	mu        sync.RWMutex

	// Опции
	autoUnload      bool // Автоматически выгружать ресурсы с RefCount = 0
	maxCacheSize    int64 // Максимальный размер кеша в байтах
	currentCacheSize int64

	// Асинхронная загрузка
	loadQueue   chan *loadRequest
	loadWorkers int
	wg          sync.WaitGroup
	running     bool
}

// loadRequest запрос на загрузку ресурса
type loadRequest struct {
	path     string
	resType  ResourceType
	callback func(ResourceID, error)
}

// NewResourceManager создает новый менеджер ресурсов
func NewResourceManager(loadWorkers int, maxCacheSize int64) *ResourceManager {
	if loadWorkers <= 0 {
		loadWorkers = 4
	}

	return &ResourceManager{
		resources:       make(map[ResourceID]*Resource),
		loaders:         make(map[ResourceType]ResourceLoader),
		cache:           make(map[string]ResourceID),
		autoUnload:      true,
		maxCacheSize:    maxCacheSize,
		currentCacheSize: 0,
		loadQueue:       make(chan *loadRequest, 100),
		loadWorkers:     loadWorkers,
		running:         false,
	}
}

// Start запускает воркеры для асинхронной загрузки
func (rm *ResourceManager) Start() {
	rm.mu.Lock()
	if rm.running {
		rm.mu.Unlock()
		return
	}

	rm.running = true
	rm.mu.Unlock()

	for i := 0; i < rm.loadWorkers; i++ {
		rm.wg.Add(1)
		go rm.loadWorker()
	}
}

// Stop останавливает воркеры загрузки
func (rm *ResourceManager) Stop() {
	rm.mu.Lock()
	if !rm.running {
		rm.mu.Unlock()
		return
	}

	rm.running = false
	close(rm.loadQueue)
	rm.mu.Unlock()

	rm.wg.Wait()
}

// loadWorker обрабатывает запросы на загрузку
func (rm *ResourceManager) loadWorker() {
	defer rm.wg.Done()

	for req := range rm.loadQueue {
		id, err := rm.LoadSync(req.path, req.resType)
		if req.callback != nil {
			req.callback(id, err)
		}
	}
}

// RegisterLoader регистрирует загрузчик для типа ресурсов
func (rm *ResourceManager) RegisterLoader(loader ResourceLoader) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.loaders[loader.GetType()] = loader
}

// LoadSync синхронно загружает ресурс
func (rm *ResourceManager) LoadSync(path string, resType ResourceType) (ResourceID, error) {
	// Проверяем кеш
	rm.mu.RLock()
	if cachedID, exists := rm.cache[path]; exists {
		if res, exists := rm.resources[cachedID]; exists && res.IsLoaded() {
			res.AddRef()
			rm.mu.RUnlock()
			return cachedID, nil
		}
	}
	rm.mu.RUnlock()

	// Получаем загрузчик
	rm.mu.RLock()
	loader, exists := rm.loaders[resType]
	rm.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("no loader registered for type %s", resType)
	}

	// Создаем ресурс
	id := ResourceID(path) // Используем path как ID
	resource := &Resource{
		ID:       id,
		Path:     path,
		Type:     resType,
		State:    ResourceStateLoading,
		RefCount: 1,
	}

	rm.mu.Lock()
	rm.resources[id] = resource
	rm.cache[path] = id
	rm.mu.Unlock()

	// Загружаем данные
	data, err := loader.Load(path)
	if err != nil {
		resource.mu.Lock()
		resource.State = ResourceStateError
		resource.Error = err
		resource.mu.Unlock()
		return id, err
	}

	// Обновляем ресурс
	resource.mu.Lock()
	resource.Data = data
	resource.State = ResourceStateLoaded
	resource.mu.Unlock()

	rm.mu.Lock()
	rm.currentCacheSize += resource.Size
	rm.mu.Unlock()

	// Проверяем лимит кеша
	rm.checkCacheSize()

	return id, nil
}

// LoadAsync асинхронно загружает ресурс
func (rm *ResourceManager) LoadAsync(path string, resType ResourceType, callback func(ResourceID, error)) {
	rm.mu.RLock()
	running := rm.running
	rm.mu.RUnlock()

	if !running {
		// Если воркеры не запущены, загружаем синхронно
		id, err := rm.LoadSync(path, resType)
		if callback != nil {
			callback(id, err)
		}
		return
	}

	req := &loadRequest{
		path:     path,
		resType:  resType,
		callback: callback,
	}

	select {
	case rm.loadQueue <- req:
	default:
		// Очередь переполнена, загружаем синхронно
		id, err := rm.LoadSync(path, resType)
		if callback != nil {
			callback(id, err)
		}
	}
}

// Get получает ресурс по ID
func (rm *ResourceManager) Get(id ResourceID) (*Resource, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resource, exists := rm.resources[id]
	if !exists {
		return nil, ErrResourceNotFound
	}

	return resource, nil
}

// Unload выгружает ресурс
func (rm *ResourceManager) Unload(id ResourceID) error {
	rm.mu.Lock()
	resource, exists := rm.resources[id]
	if !exists {
		rm.mu.Unlock()
		return ErrResourceNotFound
	}

	// Уменьшаем счетчик ссылок
	refCount := resource.Release()

	// Если есть ссылки, не выгружаем
	if refCount > 0 {
		rm.mu.Unlock()
		return nil
	}

	// Удаляем из кеша
	delete(rm.cache, resource.Path)
	delete(rm.resources, id)

	rm.currentCacheSize -= resource.Size
	rm.mu.Unlock()

	// Выгружаем данные
	loader, exists := rm.loaders[resource.Type]
	if exists && resource.Data != nil {
		return loader.Unload(resource.Data)
	}

	return nil
}

// checkCacheSize проверяет размер кеша и выгружает неиспользуемые ресурсы
func (rm *ResourceManager) checkCacheSize() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.maxCacheSize <= 0 || rm.currentCacheSize <= rm.maxCacheSize {
		return
	}

	// Собираем ресурсы с нулевым RefCount
	toUnload := make([]ResourceID, 0)
	for id, res := range rm.resources {
		if res.GetRefCount() == 0 {
			toUnload = append(toUnload, id)
		}
	}

	// Выгружаем до достижения лимита
	for _, id := range toUnload {
		if rm.currentCacheSize <= rm.maxCacheSize {
			break
		}

		resource := rm.resources[id]
		delete(rm.cache, resource.Path)
		delete(rm.resources, id)
		rm.currentCacheSize -= resource.Size

		// Выгружаем данные
		loader, exists := rm.loaders[resource.Type]
		if exists && resource.Data != nil {
			loader.Unload(resource.Data)
		}
	}
}

// Clear очищает все ресурсы
func (rm *ResourceManager) Clear() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for id, resource := range rm.resources {
		loader, exists := rm.loaders[resource.Type]
		if exists && resource.Data != nil {
			loader.Unload(resource.Data)
		}
		delete(rm.resources, id)
	}

	rm.cache = make(map[string]ResourceID)
	rm.currentCacheSize = 0
}

// GetLoadedCount возвращает количество загруженных ресурсов
func (rm *ResourceManager) GetLoadedCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.resources)
}

// GetCacheSize возвращает текущий размер кеша
func (rm *ResourceManager) GetCacheSize() int64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.currentCacheSize
}
