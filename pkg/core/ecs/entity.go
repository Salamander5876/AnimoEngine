package ecs

import (
	"sync"
	"sync/atomic"
)

// EntityID представляет уникальный идентификатор сущности
type EntityID uint64

// Entity представляет игровой объект в ECS системе
type Entity struct {
	ID         EntityID
	componentMask uint64 // Битовая маска для быстрой проверки компонентов
}

// EntityManager управляет всеми сущностями в игре
type EntityManager struct {
	nextID        uint64
	entities      map[EntityID]*Entity
	freeIDs       []EntityID // Пул освободившихся ID для переиспользования
	entityPool    sync.Pool
	mu            sync.RWMutex
	componentMgr  *ComponentManager
}

// NewEntityManager создает новый менеджер сущностей
func NewEntityManager() *EntityManager {
	em := &EntityManager{
		nextID:   1,
		entities: make(map[EntityID]*Entity),
		freeIDs:  make([]EntityID, 0),
		entityPool: sync.Pool{
			New: func() interface{} {
				return &Entity{}
			},
		},
		componentMgr: NewComponentManager(),
	}
	return em
}

// CreateEntity создает новую сущность
func (em *EntityManager) CreateEntity() EntityID {
	em.mu.Lock()
	defer em.mu.Unlock()

	var id EntityID

	// Переиспользуем освободившиеся ID если есть
	if len(em.freeIDs) > 0 {
		id = em.freeIDs[len(em.freeIDs)-1]
		em.freeIDs = em.freeIDs[:len(em.freeIDs)-1]
	} else {
		id = EntityID(atomic.AddUint64(&em.nextID, 1))
	}

	entity := em.entityPool.Get().(*Entity)
	entity.ID = id
	entity.componentMask = 0

	em.entities[id] = entity

	return id
}

// DestroyEntity удаляет сущность и все её компоненты
func (em *EntityManager) DestroyEntity(id EntityID) {
	em.mu.Lock()
	defer em.mu.Unlock()

	entity, exists := em.entities[id]
	if !exists {
		return
	}

	// Удаляем все компоненты
	em.componentMgr.RemoveAllComponents(id)

	// Очищаем и возвращаем в пул
	entity.componentMask = 0
	em.entityPool.Put(entity)

	delete(em.entities, id)
	em.freeIDs = append(em.freeIDs, id)
}

// GetEntity возвращает сущность по ID
func (em *EntityManager) GetEntity(id EntityID) (*Entity, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	entity, exists := em.entities[id]
	return entity, exists
}

// Exists проверяет, существует ли сущность
func (em *EntityManager) Exists(id EntityID) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	_, exists := em.entities[id]
	return exists
}

// GetAllEntities возвращает все активные сущности
func (em *EntityManager) GetAllEntities() []EntityID {
	em.mu.RLock()
	defer em.mu.RUnlock()

	ids := make([]EntityID, 0, len(em.entities))
	for id := range em.entities {
		ids = append(ids, id)
	}
	return ids
}

// GetEntitiesWithComponents возвращает все сущности с указанными компонентами
func (em *EntityManager) GetEntitiesWithComponents(componentMask uint64) []EntityID {
	em.mu.RLock()
	defer em.mu.RUnlock()

	result := make([]EntityID, 0)
	for id, entity := range em.entities {
		if (entity.componentMask & componentMask) == componentMask {
			result = append(result, id)
		}
	}
	return result
}

// AddComponent добавляет компонент к сущности
func (em *EntityManager) AddComponent(id EntityID, component Component) error {
	em.mu.Lock()
	entity, exists := em.entities[id]
	if !exists {
		em.mu.Unlock()
		return ErrEntityNotFound
	}
	em.mu.Unlock()

	// Добавляем компонент через ComponentManager
	if err := em.componentMgr.AddComponent(id, component); err != nil {
		return err
	}

	// Обновляем битовую маску
	em.mu.Lock()
	componentType := em.componentMgr.GetComponentType(component)
	entity.componentMask |= (1 << componentType)
	em.mu.Unlock()

	return nil
}

// RemoveComponent удаляет компонент из сущности
func (em *EntityManager) RemoveComponent(id EntityID, componentType ComponentType) error {
	em.mu.Lock()
	entity, exists := em.entities[id]
	if !exists {
		em.mu.Unlock()
		return ErrEntityNotFound
	}
	em.mu.Unlock()

	// Удаляем компонент через ComponentManager
	if err := em.componentMgr.RemoveComponent(id, componentType); err != nil {
		return err
	}

	// Обновляем битовую маску
	em.mu.Lock()
	entity.componentMask &^= (1 << componentType)
	em.mu.Unlock()

	return nil
}

// GetComponent получает компонент сущности
func (em *EntityManager) GetComponent(id EntityID, componentType ComponentType) (Component, error) {
	return em.componentMgr.GetComponent(id, componentType)
}

// HasComponent проверяет наличие компонента у сущности
func (em *EntityManager) HasComponent(id EntityID, componentType ComponentType) bool {
	em.mu.RLock()
	entity, exists := em.entities[id]
	em.mu.RUnlock()

	if !exists {
		return false
	}

	return (entity.componentMask & (1 << componentType)) != 0
}

// GetComponentManager возвращает менеджер компонентов
func (em *EntityManager) GetComponentManager() *ComponentManager {
	return em.componentMgr
}

// Count возвращает количество активных сущностей
func (em *EntityManager) Count() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return len(em.entities)
}

// Clear удаляет все сущности
func (em *EntityManager) Clear() {
	em.mu.Lock()
	defer em.mu.Unlock()

	for id := range em.entities {
		em.componentMgr.RemoveAllComponents(id)
		entity := em.entities[id]
		entity.componentMask = 0
		em.entityPool.Put(entity)
	}

	em.entities = make(map[EntityID]*Entity)
	em.freeIDs = em.freeIDs[:0]
}
