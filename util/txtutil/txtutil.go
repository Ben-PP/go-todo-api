package txtutil

import "fmt"

func AddLineNumberToFileName(fileName string, lineNumber int) string {
	return fmt.Sprintf("%v: %d", fileName, lineNumber)
}
