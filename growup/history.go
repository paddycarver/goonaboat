package main

type history struct {
	messages [][]byte
	writer chan []byte
	reader chan chan []byte
}

func (h *history) listen() {
	for {
		select {
		case msg := <-h.writer:
			h.messages = append(h.messages, msg)
		case resp := <-h.reader:
			for _, msg := range h.messages {
				resp <- msg
			}
			close(resp)
		}
	}
}

func (h *history) dump() [][]byte {
	resp := make(chan []byte)
	h.reader <- resp
	result := make([][]byte, 0)
	for msg := range resp {
		result = append(result, msg)
	}
	return result
}
