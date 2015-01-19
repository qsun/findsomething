package main

import "log"
import "bufio"
import "net"
import "golang.org/x/exp/inotify"

type Monitoring struct {
	Path           string
	SearchSockFile string
	Files          []string
	Change         chan inotify.Event
	Watcher        inotify.Watcher
}

func NewMonitoring(path string, searchSockFile string) *Monitoring {
	monitor := new(Monitoring)
	monitor.Path = path
	monitor.Change = make(chan inotify.Event)
	monitor.SearchSockFile = searchSockFile
	return monitor
}

func (w *Monitoring) Start() {
	watcher, err := inotify.NewWatcher()
	w.Watcher = *watcher
	if err != nil {
		log.Fatal("Can not initialize watcher ", err)
	}

	err = watcher.Watch(w.Path)
	watcher.AddWatch(w.Path, inotify.IN_ALL_EVENTS)
	if err != nil {
		log.Fatal("Failed to watch ", w.Path, ": ", err)
	}

	for {
		select {
		case ev := <-watcher.Event:
			// log.Println("Event: ", ev)
			w.Change <- *ev

		case err := <-watcher.Error:
			log.Println("Error: ", err)
		}
	}
}

func (w *Monitoring) StartSearch() {
	l, err := net.Listen("unix", w.SearchSockFile)
	if err != nil {
		log.Fatal("Failed to start search daemon: ", err)
	}

	defer func () {
		l.Close()
		log.Println("Failure")
	}()
	
	for {
		fd, err := l.Accept()
		if err != nil {
			log.Println("Failed to accept fd: ", err)
			continue
		}

		l, _, err := bufio.NewReader(fd).ReadLine()
		if err != nil {
			log.Println("Failed to readline: ", err)
		}

		line := string(l)

		log.Println("This line is: ", line)
	}
}

func (w *Monitoring) ProcessEvent(event inotify.Event) {
	if event.Mask&inotify.IN_CREATE == inotify.IN_CREATE {
		w.AddFile(event)
	}

	if event.Mask&inotify.IN_DELETE == inotify.IN_DELETE {
		w.RemoveFile(event)
	}

	if event.Mask&inotify.IN_MOVED_TO == inotify.IN_MOVED_TO {
		w.AddFile(event)
	}

	if event.Mask&inotify.IN_MOVED_FROM == inotify.IN_MOVED_FROM {
		w.RemoveFile(event)
	}
}

func (w *Monitoring) AddFile(event inotify.Event) {
	file := event.Name
	w.RemoveFile(event)
	w.Files = append(w.Files, file)
	log.Println("Added: ", file)

	if event.Mask&inotify.IN_ISDIR == inotify.IN_ISDIR {
		w.Watcher.AddWatch(file, inotify.IN_ALL_EVENTS)
		log.Println("Add watcher: ", file)
	}

}

func (w *Monitoring) RemoveFile(event inotify.Event) {
	file := event.Name

	if event.Mask&inotify.IN_ISDIR == inotify.IN_ISDIR {
		w.Watcher.RemoveWatch(file)
	}

	for i, f := range w.Files {
		if f == file {
			w.Files = append(w.Files[:i], w.Files[i+1:]...)
			break
		}
	}
	log.Println("Removed: ", file)
}
