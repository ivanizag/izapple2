package izapple2

import (
	"sync/atomic"
)

/*
The removable media drives are the drives a file can be loaded on while the
emulation runs, by command or by dragging a file on the window. They are
registered by the cards that own them, in the order the frontends and the
CommandLoadDisk command use to address them.
*/

// mediaName is the name of the media loaded on a removable media drive. It is
// written by the emulation goroutine when a file is inserted, and read by the
// frontends to show what each drive has, hence the atomic access.
type mediaName struct {
	value atomic.Value
}

func (m *mediaName) set(name string) {
	m.value.Store(name)
}

func (m *mediaName) get() string {
	name, _ := m.value.Load().(string)
	return name
}

// removableMediaDrive is a drive registered as a target for the files loaded
// while the emulation runs
type removableMediaDrive struct {
	label string // Slot and drive number, for example "S6D1"
	drive drive
}

func (a *Apple2) registerRemovableMediaDrive(d drive, label string) {
	a.removableMediaDrives = append(a.removableMediaDrives, removableMediaDrive{
		label: label,
		drive: d,
	})
}

// DriveInfo describes a removable media drive as seen from outside
type DriveInfo struct {
	Label string // Slot and drive number, for example "S6D1"
	Media string // Name of the media inserted, empty when the drive is empty
}

// GetRemovableMediaDrives returns the drives a file can be loaded on, in the
// order SendLoadDisk() addresses them. The frontends use it to show where a
// dragged file can be dropped and what each drive has inserted.
func (a *Apple2) GetRemovableMediaDrives() []DriveInfo {
	drives := make([]DriveInfo, len(a.removableMediaDrives))
	for i, d := range a.removableMediaDrives {
		drives[i] = DriveInfo{
			Label: d.label,
			Media: d.drive.getMediaName(),
		}
	}
	return drives
}
