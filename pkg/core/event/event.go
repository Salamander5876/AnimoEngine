package event

import (
	"sync"
	"time"
)

// EventType представляет тип события
type EventType string

// Event представляет игровое событие
type Event struct {
	Type      EventType              // Тип события
	Data      interface{}            // Данные события
	Timestamp time.Time              // Время создания события
	Priority  int                    // Приоритет (больше = выше приоритет)
	Cancelled bool                   // Флаг отмены события
	Metadata  map[string]interface{} // Дополнительные метаданные
}

// NewEvent создает новое событие
func NewEvent(eventType EventType, data interface{}) *Event {
	return &Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		Priority:  0,
		Cancelled: false,
		Metadata:  make(map[string]interface{}),
	}
}

// NewEventWithPriority создает событие с заданным приоритетом
func NewEventWithPriority(eventType EventType, data interface{}, priority int) *Event {
	event := NewEvent(eventType, data)
	event.Priority = priority
	return event
}

// Cancel отменяет событие
func (e *Event) Cancel() {
	e.Cancelled = true
}

// IsCancelled возвращает, было ли событие отменено
func (e *Event) IsCancelled() bool {
	return e.Cancelled
}

// SetMetadata устанавливает метаданные события
func (e *Event) SetMetadata(key string, value interface{}) {
	e.Metadata[key] = value
}

// GetMetadata получает метаданные события
func (e *Event) GetMetadata(key string) (interface{}, bool) {
	value, exists := e.Metadata[key]
	return value, exists
}

// EventHandler функция-обработчик события
type EventHandler func(*Event)

// EventListener представляет подписчика на события
type EventListener struct {
	ID       string
	Handler  EventHandler
	Priority int
	Once     bool // Если true, обработчик вызывается только один раз
}

// EventBus представляет шину событий для pub/sub
type EventBus struct {
	listeners map[EventType][]*EventListener
	queue     chan *Event
	workerNum int
	mu        sync.RWMutex
	wg        sync.WaitGroup
	running   bool
	nextID    uint64
}

// NewEventBus создает новую шину событий
func NewEventBus(queueSize int, workerNum int) *EventBus {
	if workerNum <= 0 {
		workerNum = 1
	}

	return &EventBus{
		listeners: make(map[EventType][]*EventListener),
		queue:     make(chan *Event, queueSize),
		workerNum: workerNum,
		running:   false,
		nextID:    0,
	}
}

// Start запускает обработку событий
func (eb *EventBus) Start() {
	eb.mu.Lock()
	if eb.running {
		eb.mu.Unlock()
		return
	}

	eb.running = true
	eb.mu.Unlock()

	// Запускаем воркеры для обработки событий
	for i := 0; i < eb.workerNum; i++ {
		eb.wg.Add(1)
		go eb.worker()
	}
}

// Stop останавливает обработку событий
func (eb *EventBus) Stop() {
	eb.mu.Lock()
	if !eb.running {
		eb.mu.Unlock()
		return
	}

	eb.running = false
	close(eb.queue)
	eb.mu.Unlock()

	eb.wg.Wait()
}

// worker обрабатывает события из очереди
func (eb *EventBus) worker() {
	defer eb.wg.Done()

	for event := range eb.queue {
		if event.IsCancelled() {
			continue
		}

		eb.processEvent(event)
	}
}

// processEvent обрабатывает одно событие
func (eb *EventBus) processEvent(event *Event) {
	eb.mu.RLock()
	listeners, exists := eb.listeners[event.Type]
	if !exists || len(listeners) == 0 {
		eb.mu.RUnlock()
		return
	}

	// Копируем список слушателей для безопасной итерации
	listenersCopy := make([]*EventListener, len(listeners))
	copy(listenersCopy, listeners)
	eb.mu.RUnlock()

	// Сортируем по приоритету (выше приоритет = раньше вызывается)
	sortListenersByPriority(listenersCopy)

	// Вызываем обработчики
	listenersToRemove := make([]string, 0)
	for _, listener := range listenersCopy {
		if event.IsCancelled() {
			break
		}

		listener.Handler(event)

		if listener.Once {
			listenersToRemove = append(listenersToRemove, listener.ID)
		}
	}

	// Удаляем одноразовые обработчики
	if len(listenersToRemove) > 0 {
		eb.mu.Lock()
		for _, id := range listenersToRemove {
			eb.removeListenerByID(event.Type, id)
		}
		eb.mu.Unlock()
	}
}

// Emit отправляет событие в очередь обработки
func (eb *EventBus) Emit(event *Event) {
	eb.mu.RLock()
	running := eb.running
	eb.mu.RUnlock()

	if !running {
		return
	}

	select {
	case eb.queue <- event:
	default:
		// Очередь переполнена, можно логировать или обработать
	}
}

// EmitSync синхронно обрабатывает событие
func (eb *EventBus) EmitSync(event *Event) {
	eb.processEvent(event)
}

// Subscribe подписывается на события заданного типа
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) string {
	return eb.SubscribeWithPriority(eventType, handler, 0)
}

// SubscribeWithPriority подписывается на события с заданным приоритетом
func (eb *EventBus) SubscribeWithPriority(eventType EventType, handler EventHandler, priority int) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	id := eb.generateID()
	listener := &EventListener{
		ID:       id,
		Handler:  handler,
		Priority: priority,
		Once:     false,
	}

	eb.listeners[eventType] = append(eb.listeners[eventType], listener)
	return id
}

// SubscribeOnce подписывается на одно событие (обработчик вызывается один раз)
func (eb *EventBus) SubscribeOnce(eventType EventType, handler EventHandler) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	id := eb.generateID()
	listener := &EventListener{
		ID:       id,
		Handler:  handler,
		Priority: 0,
		Once:     true,
	}

	eb.listeners[eventType] = append(eb.listeners[eventType], listener)
	return id
}

// Unsubscribe отписывается от событий
func (eb *EventBus) Unsubscribe(eventType EventType, listenerID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.removeListenerByID(eventType, listenerID)
}

// UnsubscribeAll отписывается от всех событий заданного типа
func (eb *EventBus) UnsubscribeAll(eventType EventType) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	delete(eb.listeners, eventType)
}

// removeListenerByID удаляет слушателя по ID (не thread-safe)
func (eb *EventBus) removeListenerByID(eventType EventType, listenerID string) {
	listeners, exists := eb.listeners[eventType]
	if !exists {
		return
	}

	for i, listener := range listeners {
		if listener.ID == listenerID {
			eb.listeners[eventType] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

// generateID генерирует уникальный ID для слушателя (не thread-safe)
func (eb *EventBus) generateID() string {
	eb.nextID++
	return string(rune(eb.nextID))
}

// HasListeners проверяет, есть ли подписчики на событие
func (eb *EventBus) HasListeners(eventType EventType) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	listeners, exists := eb.listeners[eventType]
	return exists && len(listeners) > 0
}

// ListenerCount возвращает количество подписчиков на событие
func (eb *EventBus) ListenerCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	listeners, exists := eb.listeners[eventType]
	if !exists {
		return 0
	}
	return len(listeners)
}

// Clear удаляет всех подписчиков
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.listeners = make(map[EventType][]*EventListener)
}

// sortListenersByPriority сортирует слушателей по приоритету
func sortListenersByPriority(listeners []*EventListener) {
	// Простая сортировка пузырьком (достаточно для небольшого количества слушателей)
	n := len(listeners)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if listeners[j].Priority < listeners[j+1].Priority {
				listeners[j], listeners[j+1] = listeners[j+1], listeners[j]
			}
		}
	}
}
