package command

type ParserResult struct {
	Line          [][]byte
	IsComplete    bool
	BytesConsumed int
	InPound       bool
}

func Parse(data []byte) ParserResult {
	// the bytes to return for a full line/command
	// commands are separated by newline
	var line [][]byte
	// the word to build up
	// words are separated by space
	var argument []byte
	// amount of bytes consumed
	bytesConsumed := 0
	// '#' char allow space in value (ex. #hello world# --> value: hello world)
	inPound := false

	for _, char := range data {
		if char == '#' && !inPound {
			bytesConsumed++
			inPound = true

		} else if char == '#' && inPound {
			bytesConsumed++
			inPound = false

		} else if !inPound && char == '\n' {
			// create a copy of arguments because slice point to underlying array
			// doing this avoids pointing to overwritten argument
			argCopy := append([]byte(nil), argument...)

			if argCopy == nil {
				// ensure it's empty not nil if input is also empty
				argCopy = []byte{}
			}

			line = append(line, argCopy)
			bytesConsumed++

			return ParserResult{
				Line:          line,
				IsComplete:    true,
				BytesConsumed: bytesConsumed,
				InPound:       inPound,
			}

		} else if !inPound && char == ' ' {
			// create a copy of arguments because slice point to underlying array
			// doing this avoids pointing to overwritten argument
			argCopy := append([]byte(nil), argument...)
			line = append(line, argCopy)
			argument = nil
			bytesConsumed++

		} else {
			argument = append(argument, char)
			bytesConsumed++
		}
	}

	return ParserResult{
		Line:          line,
		IsComplete:    false,
		BytesConsumed: bytesConsumed,
		InPound:       inPound,
	}
}
