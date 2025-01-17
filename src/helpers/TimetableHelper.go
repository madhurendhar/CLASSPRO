package helpers

import (
	"goscraper/src/types"

	"strings"
)

var batch1 = types.Batch{
	Batch: "1",
	Slots: []types.Slot{
		{Day: 1, DayOrder: "Day 1", Slots: []string{"A", "A", "F", "F", "G", "P6", "P7", "P8", "P9", "P10"}},
		{Day: 2, DayOrder: "Day 2", Slots: []string{"P11", "P12", "P13", "P14", "P15", "B", "B", "G", "G", "A"}},
		{Day: 3, DayOrder: "Day 3", Slots: []string{"C", "C", "A", "D", "B", "P26", "P27", "P28", "P29", "P30"}},
		{Day: 4, DayOrder: "Day 4", Slots: []string{"P31", "P32", "P33", "P34", "P35", "D", "D", "B", "E", "C"}},
		{Day: 5, DayOrder: "Day 5", Slots: []string{"E", "E", "C", "F", "D", "P46", "P47", "P48", "P49", "P50"}},
	},
}

var batch2 = types.Batch{
	Batch: "2",
	Slots: []types.Slot{
		{Day: 1, DayOrder: "Day 1", Slots: []string{"P1", "P2", "P3", "P4", "P5", "A", "A", "F", "F", "G"}},
		{Day: 2, DayOrder: "Day 2", Slots: []string{"B", "B", "G", "G", "A", "P16", "P17", "P18", "P19", "P20"}},
		{Day: 3, DayOrder: "Day 3", Slots: []string{"P21", "P22", "P23", "P24", "P25", "C", "C", "A", "D", "B"}},
		{Day: 4, DayOrder: "Day 4", Slots: []string{"D", "D", "B", "E", "C", "P36", "P37", "P38", "P39", "P40"}},
		{Day: 5, DayOrder: "Day 5", Slots: []string{"P41", "P42", "P43", "P44", "P45", "E", "E", "C", "F", "D"}},
	},
}

type Timetable struct {
	cookie string
}

func NewTimetable(cookie string) *Timetable {
	return &Timetable{cookie: cookie}
}

func (t *Timetable) GetTimetable() (*types.TimetableResult, error) {
	coursePage := NewCoursePage(t.cookie)
	courseList, err := coursePage.GetCourses()
	if err != nil {
		return nil, err
	}

	mappedSchedule := t.mapWithFallback(*courseList)
	return mappedSchedule, nil
}

func (t *Timetable) getSlotsFromRange(slotRange string) []string {
	return strings.Split(slotRange, "-")
}

func (t *Timetable) mapSlotsToSubjects(batch types.Batch, subjects []types.Course) []types.DaySchedule {
	slotMapping := make(map[string]types.TableSlot)

	for _, subject := range subjects {
		var slots []string
		if strings.Contains(subject.Slot, "-") {
			slots = t.getSlotsFromRange(subject.Slot)
		} else {
			slots = []string{subject.Slot}
		}

		isOnline := strings.Contains(strings.ToLower(subject.Room), "online")
		slotType := "Practical"
		if !isOnline {
			slotType = subject.SlotType
		}

		for _, slot := range slots {
			slotMapping[slot] = types.TableSlot{
				Code:       subject.Code,
				Name:       subject.Title,
				Online:     isOnline,
				CourseType: slotType,
				RoomNo:     subject.Room,
				Slot:       slot,
			}
		}
	}

	var schedule []types.DaySchedule
	for _, day := range batch.Slots {
		var table []interface{}
		for _, slot := range day.Slots {
			if val, ok := slotMapping[slot]; ok {
				table = append(table, val)
			} else {
				table = append(table, nil)
			}
		}
		schedule = append(schedule, types.DaySchedule{Day: day.Day, Table: table})
	}

	return schedule
}

func (t *Timetable) mapWithFallback(subjects types.CourseResponse) *types.TimetableResult {
	batches := []types.Batch{batch1, batch2}

	for _, batch := range batches {
		mappedSchedule := t.mapSlotsToSubjects(batch, subjects.Courses)

		// Check if this batch contains relevant slots
		containsRelevantSlots := false
		for _, course := range subjects.Courses {
			if strings.HasPrefix(course.Slot, "P") {
				for _, day := range batch.Slots {
					for _, slot := range day.Slots {
						if strings.Contains(course.Slot, "-") {
							courseSlots := t.getSlotsFromRange(course.Slot)
							for _, courseSlot := range courseSlots {
								if courseSlot == slot {
									containsRelevantSlots = true
									break
								}
							}
						} else if course.Slot == slot {
							containsRelevantSlots = true
							break
						}
					}
					if containsRelevantSlots {
						break
					}
				}
			}
			if containsRelevantSlots {
				break
			}
		}

		if containsRelevantSlots {
			return &types.TimetableResult{
				RegNumber: subjects.RegNumber,
				Batch:     batch.Batch,
				Schedule:  mappedSchedule,
			}
		}
	}

	return nil
}
