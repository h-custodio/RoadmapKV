package command

func Parse(data []byte) ([][]byte, bool, int) {
	// the bytes to return for a full line/command
	// commands are separated by newline
	var line [][]byte
	// the word to build up
	// words are separated by space
	var argument []byte
	// amount of bytes consumed
	bytesConsumed := 0

	for _, char := range data {
		if char == '\n' {
			line = append(line, argument)
			bytesConsumed++
			return line, true, bytesConsumed

		} else if char == ' ' {
			line = append(line, argument)
			argument = nil
			bytesConsumed++

		} else {
			argument = append(argument, char)
			bytesConsumed++
		}

	}

	return nil, false, 0
}
