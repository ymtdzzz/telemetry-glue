package output

import (
	"encoding/csv"
	"fmt"
	"os"
)

// AttributesResult represents search values output
type AttributesResult struct {
	Values []string `json:"values"`
}

// Print outputs search values result in the specified format
func (r AttributesResult) Print(format Format) error {
	switch format {
	case FormatJSON:
		return r.printJSON()
	case FormatCSV:
		return r.printCSV()
	case FormatTable, "":
		return r.printTable()
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func (r AttributesResult) printJSON() error {
	return printJSON(r)
}

func (r AttributesResult) printTable() error {
	fmt.Printf("Found %d unique values:\n", len(r.Values))
	for _, value := range r.Values {
		fmt.Printf("  %s\n", value)
	}

	return nil
}

func (r AttributesResult) printCSV() error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"value"}); err != nil {
		return err
	}

	// Write values
	for _, value := range r.Values {
		if err := writer.Write([]string{value}); err != nil {
			return err
		}
	}

	return nil
}
