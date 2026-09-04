package api

import "time"

const boardTimeFlushInterval = time.Minute

type boardTimeTracker struct {
	activeBoardIdentifier int64
	activeSince           time.Time
	stopChannel           chan struct{}
}

func (handler *Handler) flushActiveBoardTimeLocked() {
	tracker := &handler.boardTime
	if tracker.activeBoardIdentifier == 0 {
		return
	}
	elapsedSeconds := int64(time.Since(tracker.activeSince).Seconds())
	tracker.activeSince = time.Now()
	if elapsedSeconds <= 0 {
		return
	}
	board := handler.store.Board(tracker.activeBoardIdentifier)
	if board == nil {
		tracker.activeBoardIdentifier = 0
		return
	}
	board.TimeSpentSeconds += elapsedSeconds
	_ = handler.store.SaveBoard(board)
}

func (handler *Handler) SetActiveBoard(boardIdentifier int64) (response, error) {
	handler.flushActiveBoardTimeLocked()
	handler.boardTime.activeBoardIdentifier = 0
	if handler.store.Board(boardIdentifier) != nil {
		handler.boardTime.activeBoardIdentifier = boardIdentifier
		handler.boardTime.activeSince = time.Now()
	}
	return okResponse(), nil
}

func (handler *Handler) StartBoardTimeTracking() {
	if handler.boardTime.stopChannel != nil {
		return
	}
	handler.boardTime.stopChannel = make(chan struct{})
	go handler.runBoardTimeFlushLoop(handler.boardTime.stopChannel)
}

func (handler *Handler) runBoardTimeFlushLoop(stopChannel chan struct{}) {
	ticker := time.NewTicker(boardTimeFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stopChannel:
			return
		case <-ticker.C:
			handler.operationMutex.Lock()
			handler.flushActiveBoardTimeLocked()
			handler.operationMutex.Unlock()
		}
	}
}

func (handler *Handler) Close() {
	handler.operationMutex.Lock()
	defer handler.operationMutex.Unlock()
	if handler.boardTime.stopChannel != nil {
		close(handler.boardTime.stopChannel)
		handler.boardTime.stopChannel = nil
	}
	handler.flushActiveBoardTimeLocked()
}
