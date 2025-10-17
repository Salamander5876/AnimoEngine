package ecs

import (
	"errors"
	"reflect"
	"sync"
)

// ComponentType представляет тип компонента
type ComponentType uint64

// Component интерфейс для всех компонентов в ECS
type Component interface {
	// Type возвращает тип компонента
	Type() ComponentType
}

// Ошибки ECS системы
var (
	ErrEntityNotFound     = errors.New("entity not found")
	ErrComponentNotFound  = errors.New("component not found")
	ErrComponentExists    = errors.New("component already exists")
	ErrInvalidComponent   = errors.New("invalid component")
	ErrMaxComponentsLimit = errors.New("max components limit reached")
)

// ComponentManager управляет всеми компонентами в системе
type ComponentManager struct {
	// Хранилище компонентов по типу и EntityID
	components map[ComponentType]map[EntityID]Component

	// Регистрация типов компонентов
	typeRegistry map[reflect.Type]ComponentType
	nextType     ComponentType

	// Пулы компонентов для переиспользования
	componentPools map[ComponentType]*sync.Pool

	mu sync.RWMutex
}

// NewComponentManager создает новый менеджер компонентов
func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		components:     make(map[ComponentType]map[EntityID]Component),
		typeRegistry:   make(map[reflect.Type]ComponentType),
		nextType:       0,
		componentPools: make(map[ComponentType]*sync.Pool),
	}
}

// RegisterComponentType регистрирует новый тип компонента
func (cm *ComponentManager) RegisterComponentType(component Component) ComponentType {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	t := reflect.TypeOf(component)
	if existingType, exists := cm.typeRegistry[t]; exists {
		return existingType
	}

	componentType := cm.nextType
	cm.nextType++

	cm.typeRegistry[t] = componentType
	cm.components[componentType] = make(map[EntityID]Component)

	return componentType
}

// GetComponentType возвращает тип компонента
func (cm *ComponentManager) GetComponentType(component Component) ComponentType {
	cm.mu.RLock()
	t := reflect.TypeOf(component)
	componentType, exists := cm.typeRegistry[t]
	cm.mu.RUnlock()

	if !exists {
		return cm.RegisterComponentType(component)
	}

	return componentType
}

// AddComponent добавляет компонент к сущности
func (cm *ComponentManager) AddComponent(entityID EntityID, component Component) error {
	if component == nil {
		return ErrInvalidComponent
	}

	componentType := cm.GetComponentType(component)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Проверяем, не существует ли уже компонент
	if _, exists := cm.components[componentType][entityID]; exists {
		return ErrComponentExists
	}

	cm.components[componentType][entityID] = component
	return nil
}

// RemoveComponent удаляет компонент из сущности
func (cm *ComponentManager) RemoveComponent(entityID EntityID, componentType ComponentType) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	components, exists := cm.components[componentType]
	if !exists {
		return ErrComponentNotFound
	}

	if _, exists := components[entityID]; !exists {
		return ErrComponentNotFound
	}

	delete(components, entityID)
	return nil
}

// GetComponent получает компонент сущности
func (cm *ComponentManager) GetComponent(entityID EntityID, componentType ComponentType) (Component, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	components, exists := cm.components[componentType]
	if !exists {
		return nil, ErrComponentNotFound
	}

	component, exists := components[entityID]
	if !exists {
		return nil, ErrComponentNotFound
	}

	return component, nil
}

// HasComponent проверяет наличие компонента у сущности
func (cm *ComponentManager) HasComponent(entityID EntityID, componentType ComponentType) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	components, exists := cm.components[componentType]
	if !exists {
		return false
	}

	_, exists = components[entityID]
	return exists
}

// GetAllComponents возвращает все компоненты сущности
func (cm *ComponentManager) GetAllComponents(entityID EntityID) []Component {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]Component, 0)
	for _, components := range cm.components {
		if component, exists := components[entityID]; exists {
			result = append(result, component)
		}
	}
	return result
}

// RemoveAllComponents удаляет все компоненты сущности
func (cm *ComponentManager) RemoveAllComponents(entityID EntityID) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, components := range cm.components {
		delete(components, entityID)
	}
}

// GetEntitiesWithComponent возвращает все сущности с заданным компонентом
func (cm *ComponentManager) GetEntitiesWithComponent(componentType ComponentType) []EntityID {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	components, exists := cm.components[componentType]
	if !exists {
		return []EntityID{}
	}

	result := make([]EntityID, 0, len(components))
	for entityID := range components {
		result = append(result, entityID)
	}
	return result
}

// GetComponentCount возвращает количество компонентов заданного типа
func (cm *ComponentManager) GetComponentCount(componentType ComponentType) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	components, exists := cm.components[componentType]
	if !exists {
		return 0
	}

	return len(components)
}

// Clear удаляет все компоненты
func (cm *ComponentManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for componentType := range cm.components {
		cm.components[componentType] = make(map[EntityID]Component)
	}
}
