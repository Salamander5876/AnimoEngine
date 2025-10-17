package ecs

import (
	"sort"
	"sync"
)

// System интерфейс для всех систем в ECS
type System interface {
	// Update вызывается каждый кадр для обновления системы
	Update(deltaTime float32, em *EntityManager)

	// Priority возвращает приоритет системы (меньше = раньше выполняется)
	Priority() int

	// Enabled возвращает, активна ли система
	Enabled() bool

	// SetEnabled устанавливает состояние системы
	SetEnabled(enabled bool)
}

// BaseSystem базовая реализация системы
type BaseSystem struct {
	priority int
	enabled  bool
}

// NewBaseSystem создает новую базовую систему
func NewBaseSystem(priority int) BaseSystem {
	return BaseSystem{
		priority: priority,
		enabled:  true,
	}
}

// Priority возвращает приоритет системы
func (s *BaseSystem) Priority() int {
	return s.priority
}

// Enabled возвращает состояние системы
func (s *BaseSystem) Enabled() bool {
	return s.enabled
}

// SetEnabled устанавливает состояние системы
func (s *BaseSystem) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// SystemManager управляет всеми системами
type SystemManager struct {
	systems []System
	mu      sync.RWMutex
}

// NewSystemManager создает новый менеджер систем
func NewSystemManager() *SystemManager {
	return &SystemManager{
		systems: make([]System, 0),
	}
}

// AddSystem добавляет систему в менеджер
func (sm *SystemManager) AddSystem(system System) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.systems = append(sm.systems, system)

	// Сортируем системы по приоритету
	sort.Slice(sm.systems, func(i, j int) bool {
		return sm.systems[i].Priority() < sm.systems[j].Priority()
	})
}

// RemoveSystem удаляет систему из менеджера
func (sm *SystemManager) RemoveSystem(system System) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, s := range sm.systems {
		if s == system {
			sm.systems = append(sm.systems[:i], sm.systems[i+1:]...)
			break
		}
	}
}

// Update обновляет все активные системы
func (sm *SystemManager) Update(deltaTime float32, em *EntityManager) {
	sm.mu.RLock()
	systems := make([]System, len(sm.systems))
	copy(systems, sm.systems)
	sm.mu.RUnlock()

	for _, system := range systems {
		if system.Enabled() {
			system.Update(deltaTime, em)
		}
	}
}

// GetSystems возвращает все системы
func (sm *SystemManager) GetSystems() []System {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	systems := make([]System, len(sm.systems))
	copy(systems, sm.systems)
	return systems
}

// Clear удаляет все системы
func (sm *SystemManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.systems = make([]System, 0)
}

// Archetype представляет группу сущностей с одинаковым набором компонентов
type Archetype struct {
	componentMask uint64
	entities      []EntityID
	mu            sync.RWMutex
}

// NewArchetype создает новый архетип
func NewArchetype(componentMask uint64) *Archetype {
	return &Archetype{
		componentMask: componentMask,
		entities:      make([]EntityID, 0),
	}
}

// AddEntity добавляет сущность в архетип
func (a *Archetype) AddEntity(entityID EntityID) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.entities = append(a.entities, entityID)
}

// RemoveEntity удаляет сущность из архетипа
func (a *Archetype) RemoveEntity(entityID EntityID) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, id := range a.entities {
		if id == entityID {
			a.entities = append(a.entities[:i], a.entities[i+1:]...)
			break
		}
	}
}

// GetEntities возвращает все сущности архетипа
func (a *Archetype) GetEntities() []EntityID {
	a.mu.RLock()
	defer a.mu.RUnlock()

	entities := make([]EntityID, len(a.entities))
	copy(entities, a.entities)
	return entities
}

// Count возвращает количество сущностей в архетипе
func (a *Archetype) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.entities)
}

// Matches проверяет, соответствует ли сущность архетипу
func (a *Archetype) Matches(entityMask uint64) bool {
	return (entityMask & a.componentMask) == a.componentMask
}

// ArchetypeManager управляет архетипами для оптимизации запросов
type ArchetypeManager struct {
	archetypes map[uint64]*Archetype
	mu         sync.RWMutex
}

// NewArchetypeManager создает новый менеджер архетипов
func NewArchetypeManager() *ArchetypeManager {
	return &ArchetypeManager{
		archetypes: make(map[uint64]*Archetype),
	}
}

// GetOrCreateArchetype получает или создает архетип для заданной маски компонентов
func (am *ArchetypeManager) GetOrCreateArchetype(componentMask uint64) *Archetype {
	am.mu.Lock()
	defer am.mu.Unlock()

	if archetype, exists := am.archetypes[componentMask]; exists {
		return archetype
	}

	archetype := NewArchetype(componentMask)
	am.archetypes[componentMask] = archetype
	return archetype
}

// FindArchetypes находит все архетипы, соответствующие заданной маске компонентов
func (am *ArchetypeManager) FindArchetypes(componentMask uint64) []*Archetype {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make([]*Archetype, 0)
	for _, archetype := range am.archetypes {
		if archetype.Matches(componentMask) {
			result = append(result, archetype)
		}
	}
	return result
}

// Clear удаляет все архетипы
func (am *ArchetypeManager) Clear() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.archetypes = make(map[uint64]*Archetype)
}
