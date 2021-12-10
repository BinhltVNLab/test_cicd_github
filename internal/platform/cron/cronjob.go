package cron

import (
	"time"

	"github.com/robfig/cron/v3"
)

// EtrCron : struct cron
type EtrCron struct {
	Cron    *cron.Cron
	Entries []Entry
}

// Entry : struct cron
type Entry struct {
	ID   cron.EntryID
	Name string
}

// InitCron : Init cron
func (etrCron *EtrCron) InitCron(location string) {
	loc, _ := time.LoadLocation(location)
	cron := cron.New(cron.WithLocation(loc))
	etrCron.Cron = cron
	etrCron.Entries = []Entry{}
}

// AddFuncCron : Add func cron
func (etrCron *EtrCron) AddFuncCron(spec string, name string, cmd func()) (cron.EntryID, error) {
	entryID, err := etrCron.Cron.AddFunc(spec, cmd)
	entry := Entry{
		ID:   entryID,
		Name: name,
	}
	etrCron.Entries = append(etrCron.Entries, entry)
	return entryID, err
}

// GetEntries : Get entries cron
func (etrCron *EtrCron) GetEntries() []Entry {
	return etrCron.Entries
}

// RemoveCron : Remove cron
func (etrCron *EtrCron) RemoveCron(ID cron.EntryID) {
	etrCron.Cron.Remove(ID)
	etrCron.Entries = removeEntry(ID, etrCron.Entries)
}

// StartCron : Start Cron
func (etrCron *EtrCron) StartCron() {
	etrCron.Cron.Start()
}

// StopCron : Stop Cron
func (etrCron *EtrCron) StopCron() {
	etrCron.Cron.Stop()
}

// removeEntry : Remove entry
func removeEntry(ID cron.EntryID, entries []Entry) []Entry {
	for idx, element := range entries {
		if element.ID == ID {
			return append(entries[0:idx], entries[idx+1:]...)
		}
	}
	return entries
}
