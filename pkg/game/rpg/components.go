package rpg

import (
	"github.com/Salamander5876/AnimoEngine/pkg/core/ecs"
)

// ComponentType для RPG компонентов
const (
	HealthComponentType ecs.ComponentType = iota + 100
	ManaComponentType
	StaminaComponentType
	StatsComponentType
	InventoryComponentType
	EquipmentComponentType
	QuestLogComponentType
)

// HealthComponent компонент здоровья
type HealthComponent struct {
	Current      float32
	Max          float32
	Regeneration float32 // HP в секунду
}

func (h *HealthComponent) Type() ecs.ComponentType {
	return HealthComponentType
}

// IsDead проверяет, мертв ли персонаж
func (h *HealthComponent) IsDead() bool {
	return h.Current <= 0
}

// Heal восстанавливает здоровье
func (h *HealthComponent) Heal(amount float32) {
	h.Current += amount
	if h.Current > h.Max {
		h.Current = h.Max
	}
}

// Damage наносит урон
func (h *HealthComponent) Damage(amount float32) {
	h.Current -= amount
	if h.Current < 0 {
		h.Current = 0
	}
}

// GetHealthPercent возвращает процент здоровья (0-1)
func (h *HealthComponent) GetHealthPercent() float32 {
	if h.Max == 0 {
		return 0
	}
	return h.Current / h.Max
}

// ManaComponent компонент маны
type ManaComponent struct {
	Current      float32
	Max          float32
	Regeneration float32 // Мана в секунду
}

func (m *ManaComponent) Type() ecs.ComponentType {
	return ManaComponentType
}

// HasEnoughMana проверяет наличие достаточного количества маны
func (m *ManaComponent) HasEnoughMana(cost float32) bool {
	return m.Current >= cost
}

// UseMana расходует ману
func (m *ManaComponent) UseMana(amount float32) bool {
	if !m.HasEnoughMana(amount) {
		return false
	}
	m.Current -= amount
	return true
}

// RestoreMana восстанавливает ману
func (m *ManaComponent) RestoreMana(amount float32) {
	m.Current += amount
	if m.Current > m.Max {
		m.Current = m.Max
	}
}

// StaminaComponent компонент выносливости
type StaminaComponent struct {
	Current      float32
	Max          float32
	Regeneration float32 // Выносливость в секунду
}

func (s *StaminaComponent) Type() ecs.ComponentType {
	return StaminaComponentType
}

// HasEnoughStamina проверяет наличие достаточной выносливости
func (s *StaminaComponent) HasEnoughStamina(cost float32) bool {
	return s.Current >= cost
}

// UseStamina расходует выносливость
func (s *StaminaComponent) UseStamina(amount float32) bool {
	if !s.HasEnoughStamina(amount) {
		return false
	}
	s.Current -= amount
	return true
}

// RestoreStamina восстанавливает выносливость
func (s *StaminaComponent) RestoreStamina(amount float32) {
	s.Current += amount
	if s.Current > s.Max {
		s.Current = s.Max
	}
}

// StatsComponent базовые характеристики персонажа
type StatsComponent struct {
	Level int

	// Базовые характеристики
	Strength     int // Сила (физический урон, грузоподъемность)
	Agility      int // Ловкость (скорость атаки, уклонение)
	Intelligence int // Интеллект (магический урон, мана)
	Vitality     int // Живучесть (здоровье, регенерация)
	Luck         int // Удача (шанс критического удара, лут)

	// Опыт
	Experience    int
	ExperienceToNextLevel int
}

func (s *StatsComponent) Type() ecs.ComponentType {
	return StatsComponentType
}

// AddExperience добавляет опыт и проверяет повышение уровня
func (s *StatsComponent) AddExperience(amount int) bool {
	s.Experience += amount
	if s.Experience >= s.ExperienceToNextLevel {
		s.LevelUp()
		return true
	}
	return false
}

// LevelUp повышает уровень
func (s *StatsComponent) LevelUp() {
	s.Level++
	s.Experience -= s.ExperienceToNextLevel
	s.ExperienceToNextLevel = int(float32(s.ExperienceToNextLevel) * 1.5)

	// Автоматическое повышение характеристик
	s.Strength += 2
	s.Agility += 2
	s.Intelligence += 2
	s.Vitality += 3
	s.Luck += 1
}

// GetPhysicalDamage возвращает физический урон
func (s *StatsComponent) GetPhysicalDamage() float32 {
	return float32(s.Strength) * 2.5
}

// GetMagicalDamage возвращает магический урон
func (s *StatsComponent) GetMagicalDamage() float32 {
	return float32(s.Intelligence) * 3.0
}

// GetCriticalChance возвращает шанс критического удара (0-1)
func (s *StatsComponent) GetCriticalChance() float32 {
	baseChance := 0.05
	luckBonus := float32(s.Luck) * 0.01
	return float32(baseChance) + luckBonus
}

// ItemSlot слот предмета в инвентаре
type ItemSlot struct {
	ItemID   string
	Quantity int
}

// InventoryComponent компонент инвентаря
type InventoryComponent struct {
	Slots       []ItemSlot
	MaxSlots    int
	Gold        int
	MaxWeight   float32
	CurrentWeight float32
}

func (i *InventoryComponent) Type() ecs.ComponentType {
	return InventoryComponentType
}

// AddItem добавляет предмет в инвентарь
func (i *InventoryComponent) AddItem(itemID string, quantity int) bool {
	// Ищем существующий стек
	for idx := range i.Slots {
		if i.Slots[idx].ItemID == itemID {
			i.Slots[idx].Quantity += quantity
			return true
		}
	}

	// Добавляем новый слот
	if len(i.Slots) >= i.MaxSlots {
		return false // Инвентарь полон
	}

	i.Slots = append(i.Slots, ItemSlot{
		ItemID:   itemID,
		Quantity: quantity,
	})
	return true
}

// RemoveItem удаляет предмет из инвентаря
func (i *InventoryComponent) RemoveItem(itemID string, quantity int) bool {
	for idx := range i.Slots {
		if i.Slots[idx].ItemID == itemID {
			if i.Slots[idx].Quantity < quantity {
				return false
			}

			i.Slots[idx].Quantity -= quantity
			if i.Slots[idx].Quantity == 0 {
				// Удаляем пустой слот
				i.Slots = append(i.Slots[:idx], i.Slots[idx+1:]...)
			}
			return true
		}
	}
	return false
}

// HasItem проверяет наличие предмета
func (i *InventoryComponent) HasItem(itemID string, quantity int) bool {
	for _, slot := range i.Slots {
		if slot.ItemID == itemID && slot.Quantity >= quantity {
			return true
		}
	}
	return false
}

// GetItemCount возвращает количество предмета
func (i *InventoryComponent) GetItemCount(itemID string) int {
	for _, slot := range i.Slots {
		if slot.ItemID == itemID {
			return slot.Quantity
		}
	}
	return 0
}

// EquipmentSlot тип слота экипировки
type EquipmentSlot string

const (
	SlotHead      EquipmentSlot = "head"
	SlotChest     EquipmentSlot = "chest"
	SlotLegs      EquipmentSlot = "legs"
	SlotFeet      EquipmentSlot = "feet"
	SlotMainHand  EquipmentSlot = "main_hand"
	SlotOffHand   EquipmentSlot = "off_hand"
	SlotAccessory EquipmentSlot = "accessory"
)

// EquipmentComponent компонент экипировки
type EquipmentComponent struct {
	Slots map[EquipmentSlot]string // slot -> itemID
}

func (e *EquipmentComponent) Type() ecs.ComponentType {
	return EquipmentComponentType
}

// Equip надевает предмет в слот
func (e *EquipmentComponent) Equip(slot EquipmentSlot, itemID string) string {
	if e.Slots == nil {
		e.Slots = make(map[EquipmentSlot]string)
	}

	previousItem := e.Slots[slot]
	e.Slots[slot] = itemID
	return previousItem
}

// Unequip снимает предмет со слота
func (e *EquipmentComponent) Unequip(slot EquipmentSlot) string {
	if e.Slots == nil {
		return ""
	}

	itemID := e.Slots[slot]
	delete(e.Slots, slot)
	return itemID
}

// GetEquipped возвращает предмет в слоте
func (e *EquipmentComponent) GetEquipped(slot EquipmentSlot) string {
	if e.Slots == nil {
		return ""
	}
	return e.Slots[slot]
}

// QuestStatus статус квеста
type QuestStatus int

const (
	QuestStatusNotStarted QuestStatus = iota
	QuestStatusInProgress
	QuestStatusCompleted
	QuestStatusFailed
)

// Quest структура квеста
type Quest struct {
	ID          string
	Status      QuestStatus
	Objectives  map[string]int // objective_id -> current_count
	Rewards     []string       // item_ids
	GoldReward  int
	ExpReward   int
}

// QuestLogComponent компонент журнала квестов
type QuestLogComponent struct {
	ActiveQuests    []Quest
	CompletedQuests []string // quest_ids
	FailedQuests    []string // quest_ids
}

func (q *QuestLogComponent) Type() ecs.ComponentType {
	return QuestLogComponentType
}

// StartQuest начинает квест
func (q *QuestLogComponent) StartQuest(quest Quest) {
	quest.Status = QuestStatusInProgress
	q.ActiveQuests = append(q.ActiveQuests, quest)
}

// CompleteQuest завершает квест
func (q *QuestLogComponent) CompleteQuest(questID string) bool {
	for idx, quest := range q.ActiveQuests {
		if quest.ID == questID {
			quest.Status = QuestStatusCompleted
			q.CompletedQuests = append(q.CompletedQuests, questID)
			q.ActiveQuests = append(q.ActiveQuests[:idx], q.ActiveQuests[idx+1:]...)
			return true
		}
	}
	return false
}

// UpdateObjective обновляет прогресс цели квеста
func (q *QuestLogComponent) UpdateObjective(questID, objectiveID string, count int) {
	for idx := range q.ActiveQuests {
		if q.ActiveQuests[idx].ID == questID {
			if q.ActiveQuests[idx].Objectives == nil {
				q.ActiveQuests[idx].Objectives = make(map[string]int)
			}
			q.ActiveQuests[idx].Objectives[objectiveID] = count
			break
		}
	}
}

// GetQuest возвращает квест по ID
func (q *QuestLogComponent) GetQuest(questID string) *Quest {
	for idx := range q.ActiveQuests {
		if q.ActiveQuests[idx].ID == questID {
			return &q.ActiveQuests[idx]
		}
	}
	return nil
}

// HasCompletedQuest проверяет, выполнен ли квест
func (q *QuestLogComponent) HasCompletedQuest(questID string) bool {
	for _, id := range q.CompletedQuests {
		if id == questID {
			return true
		}
	}
	return false
}
