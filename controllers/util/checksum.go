package util

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	v1 "k8s.io/api/core/v1"
)

// Calculate SHA256 from the Secret data. For simplicity, mashal the entire Data
// map into a json string and calculate the hash from that. Includes the
// object's labels as a way to give users change the checksum even when the
// certificate data itself did not change. This could be useful when users want
// to force trigger cert-watch to react to test or rerun actions.
func SecretDataChecksum(s *v1.Secret) (string, error) {
	dataJson, err := json.Marshal(s.Data)
	if err != nil {
		return "", err
	}
	labelsJson, err := json.Marshal(s.ObjectMeta.Labels)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	hash.Write(dataJson)
	hash.Write(labelsJson)
	return base64.URLEncoding.EncodeToString(hash.Sum(nil)), nil
}
