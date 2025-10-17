package ecs

import (
	"sync"
)

// World представляет игровой мир, содержащий все сущности и системы
type World struct {
	entityManager    *EntityManager
	systemManager    *SystemManager
	archetypeManager *ArchetypeManager

	// Состояние мира
	running bool
	paused  bool
	mu      sync.RWMutex
}

// NewWorld создает новый игровой мир
func NewWorld() *World {
	return &World{
		entityManager:    NewEntityManager(),
		systemManager:    NewSystemManager(),
		archetypeManager: NewArchetypeManager(),
		running:          false,
		paused:           false,
	}
}

// CreateEntity создает новую сущность в мире
func (w *World) CreateEntity() EntityID {
	return w.entityManager.CreateEntity()
}

// DestroyEntity удаляет сущность из мира
func (w *World) DestroyEntity(id EntityID) {
	w.entityManager.DestroyEntity(id)
}

// AddComponent добавляет компонент к сущности
func (w *World) AddComponent(entityID EntityID, component Component) error {
	return w.entityManager.AddComponent(entityID, component)
}

// RemoveComponent удаляет компонент из сущности
func (w *World) RemoveComponent(entityID EntityID, componentType ComponentType) error {
	return w.entityManager.RemoveComponent(entityID, componentType)
}

// GetComponent получает компонент сущности
func (w *World) GetComponent(entityID EntityID, componentType ComponentType) (Component, error) {
	return w.entityManager.GetComponent(entityID, componentType)
}

// HasComponent проверяет наличие компонента у сущности
func (w *World) HasComponent(entityID EntityID, componentType ComponentType) bool {
	return w.entityManager.HasComponent(entityID, componentType)
}

// GetAllEntities возвращает все активные сущности
func (w *World) GetAllEntities() []EntityID {
	return w.entityManager.GetAllEntities()
}

// GetEntitiesWithComponents возвращает все сущности с указанными компонентами
func (w *World) GetEntitiesWithComponents(componentMask uint64) []EntityID {
	return w.entityManager.GetEntitiesWithComponents(componentMask)
}

// AddSystem добавляет систему в мир
func (w *World) AddSystem(system System) {
	w.systemManager.AddSystem(system)
}

// RemoveSystem удаляет систему из мира
func (w *World) RemoveSystem(system System) {
	w.systemManager.RemoveSystem(system)
}

// Update обновляет все системы мира
func (w *World) Update(deltaTime float32) {
	w.mu.RLock()
	if !w.running || w.paused {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	w.systemManager.Update(deltaTime, w.entityManager)
}

// Start запускает мир
func (w *World) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.running = true
	w.paused = false
}

// Stop останавливает мир
func (w *World) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.running = false
}

// Pause приостанавливает обновление мира
func (w *World) Pause() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.paused = true
}

// Resume возобновляет обновление мира
func (w *World) Resume() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.paused = false
}

// IsRunning возвращает состояние мира
func (w *World) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.running
}

// IsPaused возвращает, приостановлен ли мир
func (w *World) IsPaused() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.paused
}

// GetEntityManager возвращает менеджер сущностей
func (w *World) GetEntityManager() *EntityManager {
	return w.entityManager
}

// GetSystemManager возвращает менеджер систем
func (w *World) GetSystemManager() *SystemManager {
	return w.systemManager
}

// GetArchetypeManager возвращает менеджер архетипов
func (w *World) GetArchetypeManager() *ArchetypeManager {
	return w.archetypeManager
}

// Clear очищает мир от всех сущностей и компонентов
func (w *World) Clear() {
	w.entityManager.Clear()
	w.archetypeManager.Clear()
}

// Destroy полностью уничтожает мир
func (w *World) Destroy() {
	w.Stop()
	w.systemManager.Clear()
	w.Clear()
}

// EntityCount возвращает количество активных сущностей
func (w *World) EntityCount() int {
	return w.entityManager.Count()
}

// Query создает запрос для поиска сущностей с определенными компонентами
type Query struct {
	world         *World
	componentMask uint64
}

// NewQuery создает новый запрос
func (w *World) NewQuery() *Query {
	return &Query{
		world:         w,
		componentMask: 0,
	}
}

// With добавляет требуемый компонент в запрос
func (q *Query) With(componentType ComponentType) *Query {
	q.componentMask |= (1 << componentType)
	return q
}

// Execute выполняет запрос и возвращает подходящие сущности
func (q *Query) Execute() []EntityID {
	return q.world.GetEntitiesWithComponents(q.componentMask)
}
