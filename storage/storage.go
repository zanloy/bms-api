package storage // import github.com/zanloy/bms-api/storage

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zanloy/bms-api/config"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
)

const SecretName = "bms-reports"

var (
	b64   = base64.StdEncoding
	mutex = sync.Mutex{}
)

func GetReport(date time.Time) (models.Report, error) {
	reports, err := loadReports()
	if err != nil {
		return models.Report{}, err
	}

	for _, report := range reports {
		if report.Date == date {
			return report, nil
		}
	}

	return models.Report{}, fmt.Errorf("Report not found.")
}

func GetReports() ([]models.Report, error) {
	return loadReports()
}

func ListReports() ([]models.ReportSummary, error) {
	reports, err := loadReports()
	if err != nil {
		return []models.ReportSummary{}, fmt.Errorf("Error while trying to load reports: %w", err)
	}

	summaries := make([]models.ReportSummary, len(reports))
	for idx, report := range reports {
		summaries[idx] = report.Summary()
	}

	return summaries, nil
}

func SaveReport(report models.Report) error {
	mutex.Lock()
	defer mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get secret
	secret, err := kubernetes.Clientset.CoreV1().Secrets(config.Namespace()).Get(ctx, SecretName, metav1.GetOptions{})
	if err != nil {
		if k8errors.IsNotFound(err) {
			var reports = []models.Report{report}
			reportBytes, err := encodeReports(reports)
			if err != nil {
				// This should never happen. Good luck getting 100% code coverage.
				return fmt.Errorf("Error while trying to encode reports: %w", err)
			}

			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: SecretName,
				},
				Data: map[string][]byte{"reports": reportBytes},
			}

			if _, err := kubernetes.Clientset.CoreV1().Secrets(config.Namespace()).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("Failed to save new reports secret: %w", err)
			}

			return nil // We're done.
		} else {
			return fmt.Errorf("Failed to pull reports from kubernetes: %w", err)
		}
	}

	var reports []models.Report

	if data, ok := secret.Data["reports"]; ok {
		reports, err = decodeReports(data)
		if err != nil {
			return fmt.Errorf("Failed to decode reports from secret: %w", err)
		}
	} else {
		return fmt.Errorf("Failed to load reports: No 'reports' field in secret [%s].", SecretName)
	}

	reports = append(reports, report)
	reports, _ = pruneReports(reports)

	reportBytes, err := encodeReports(reports)
	if err != nil {
		return fmt.Errorf("Failed to save reports: Error while trying to encode reports: %w", err)
	}
	secret.Data["reports"] = reportBytes

	_, err = kubernetes.Clientset.CoreV1().Secrets(config.Namespace()).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to save reports: Error while trying to update k8 secret: %w", err)
	}

	return nil
}

/* Private funcs */

func decodeReports(data []byte) ([]models.Report, error) {
	// gunzip data
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return []models.Report{}, err
	}
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return []models.Report{}, err
	}

	var reports []models.Report
	// unmarshal reports object bytes
	if err := json.Unmarshal(b, &reports); err != nil {
		return []models.Report{}, err
	}

	return reports, nil
}

func encodeReports(reports []models.Report) ([]byte, error) {
	b, err := json.Marshal(reports)
	if err != nil {
		return []byte{}, err
	}

	var buffer bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buffer, gzip.BestCompression)
	if err != nil {
		return []byte{}, err
	}

	if _, err = writer.Write(b); err != nil {
		return []byte{}, err
	}
	writer.Close()

	return buffer.Bytes(), nil
}

func loadReports() ([]models.Report, error) {
	mutex.Lock()
	defer mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get secret
	secret, err := kubernetes.Clientset.CoreV1().Secrets(config.Namespace()).Get(ctx, SecretName, metav1.GetOptions{})
	if err != nil {
		return []models.Report{}, err
	}

	if b, ok := secret.Data["reports"]; ok {
		reports, err := decodeReports(b)
		if err != nil {
			return []models.Report{}, fmt.Errorf("Failed to load reports: %w", err)
		}

		return reports, nil
	} else {
		return []models.Report{}, fmt.Errorf("Failed to load reports: No 'reports' field in reports secret.")
	}
}

func pruneReports(reports []models.Report) ([]models.Report, bool) {
	var max = config.MaxReports()

	if diff := len(reports) - max; diff > 0 {
		reports = reports[diff:]
		return reports, true
	}

	return reports, false
}
