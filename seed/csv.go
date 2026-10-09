package seed

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Counts struct {
	Customers     int64
	Contacts      int64
	Opportunities int64
	Activities    int64
}

type csvOutput struct {
	name   string
	file   *os.File
	writer *csv.Writer
}

// WriteCSV writes one PostgreSQL COPY-compatible CSV file per CRM table.
func WriteCSV(outputDirectory string, generator *Generator) (counts Counts, err error) {
	if generator == nil {
		return Counts{}, fmt.Errorf("generator is required")
	}
	if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
		return Counts{}, fmt.Errorf("create output directory: %w", err)
	}

	outputs := make([]*csvOutput, 0, 4)
	defer func() {
		for _, output := range outputs {
			if closeErr := output.close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}
	}()

	customers, err := openCSV(outputDirectory, "clientes.csv", []string{
		"id", "nombre", "email", "estado", "creado_en",
	})
	if err != nil {
		return Counts{}, err
	}
	outputs = append(outputs, customers)

	contacts, err := openCSV(outputDirectory, "contactos.csv", []string{
		"id", "cliente_id", "nombre", "email", "cargo", "creado_en",
	})
	if err != nil {
		return Counts{}, err
	}
	outputs = append(outputs, contacts)

	opportunities, err := openCSV(outputDirectory, "oportunidades.csv", []string{
		"id", "cliente_id", "nombre", "monto", "estado", "cierre_estimado", "creado_en",
	})
	if err != nil {
		return Counts{}, err
	}
	outputs = append(outputs, opportunities)

	activities, err := openCSV(outputDirectory, "actividades.csv", []string{
		"id", "cliente_id", "tipo", "descripcion", "ocurrio_en",
	})
	if err != nil {
		return Counts{}, err
	}
	outputs = append(outputs, activities)

	for customerID := int64(1); customerID <= generator.options.Customers; customerID++ {
		bundle, generateErr := generator.Customer(customerID)
		if generateErr != nil {
			return Counts{}, generateErr
		}

		customer := bundle.Customer
		if err := customers.write([]string{
			strconv.FormatInt(customer.ID, 10),
			customer.Name,
			customer.Email,
			customer.Status,
			formatTimestamp(customer.CreatedAt),
		}); err != nil {
			return Counts{}, err
		}
		counts.Customers++

		for _, contact := range bundle.Contacts {
			if err := contacts.write([]string{
				strconv.FormatInt(contact.ID, 10),
				strconv.FormatInt(contact.CustomerID, 10),
				contact.Name,
				contact.Email,
				contact.Role,
				formatTimestamp(contact.CreatedAt),
			}); err != nil {
				return Counts{}, err
			}
			counts.Contacts++
		}

		for _, opportunity := range bundle.Opportunities {
			if err := opportunities.write([]string{
				strconv.FormatInt(opportunity.ID, 10),
				strconv.FormatInt(opportunity.CustomerID, 10),
				opportunity.Name,
				opportunity.Amount,
				opportunity.Status,
				opportunity.CloseDate.Format(time.DateOnly),
				formatTimestamp(opportunity.CreatedAt),
			}); err != nil {
				return Counts{}, err
			}
			counts.Opportunities++
		}

		for _, activity := range bundle.Activities {
			if err := activities.write([]string{
				strconv.FormatInt(activity.ID, 10),
				strconv.FormatInt(activity.CustomerID, 10),
				activity.Type,
				activity.Description,
				formatTimestamp(activity.OccurredAt),
			}); err != nil {
				return Counts{}, err
			}
			counts.Activities++
		}
	}

	return counts, nil
}

func openCSV(directory, name string, header []string) (*csvOutput, error) {
	file, err := os.Create(filepath.Join(directory, name))
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", name, err)
	}

	output := &csvOutput{name: name, file: file, writer: csv.NewWriter(file)}
	if err := output.writer.Write(header); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("write %s header: %w", name, err)
	}
	return output, nil
}

func (o *csvOutput) write(row []string) error {
	if err := o.writer.Write(row); err != nil {
		return fmt.Errorf("write %s: %w", o.name, err)
	}
	return nil
}

func (o *csvOutput) close() error {
	o.writer.Flush()
	if err := o.writer.Error(); err != nil {
		_ = o.file.Close()
		return fmt.Errorf("flush %s: %w", o.name, err)
	}
	if err := o.file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", o.name, err)
	}
	return nil
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
