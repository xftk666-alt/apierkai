package service

import (
	"context"
	"errors"
	"testing"
)

type stubCommerceSettingRepo struct {
	values map[string]string
}

func newStubCommerceSettingRepo(values map[string]string) *stubCommerceSettingRepo {
	if values == nil {
		values = make(map[string]string)
	}
	return &stubCommerceSettingRepo{values: values}
}

func (r *stubCommerceSettingRepo) Get(_ context.Context, key string) (*Setting, error) {
	if value, ok := r.values[key]; ok {
		return &Setting{Key: key, Value: value}, nil
	}
	return nil, ErrSettingNotFound
}

func (r *stubCommerceSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := r.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (r *stubCommerceSettingRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}

func (r *stubCommerceSettingRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (r *stubCommerceSettingRepo) SetMultiple(_ context.Context, settings map[string]string) error {
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *stubCommerceSettingRepo) GetAll(_ context.Context) (map[string]string, error) {
	result := make(map[string]string, len(r.values))
	for key, value := range r.values {
		result[key] = value
	}
	return result, nil
}

func (r *stubCommerceSettingRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

func TestCommerceFeatureServiceValidateCallbackSecret(t *testing.T) {
	t.Parallel()

	repo := newStubCommerceSettingRepo(map[string]string{
		SettingKeyCommerceCallbackSecret: "shared-secret",
	})
	settingService := NewSettingService(repo, nil)
	featureService := NewCommerceFeatureService(settingService)

	if err := featureService.ValidateCallbackSecret(context.Background(), "shared-secret"); err != nil {
		t.Fatalf("ValidateCallbackSecret() unexpected error: %v", err)
	}

	err := featureService.ValidateCallbackSecret(context.Background(), "wrong-secret")
	if !errors.Is(err, ErrCommerceCallbackUnauthorized) {
		t.Fatalf("ValidateCallbackSecret() error = %v, want %v", err, ErrCommerceCallbackUnauthorized)
	}
}

func TestCommerceFeatureServiceValidateCallbackSecretNotConfigured(t *testing.T) {
	t.Parallel()

	repo := newStubCommerceSettingRepo(nil)
	settingService := NewSettingService(repo, nil)
	featureService := NewCommerceFeatureService(settingService)

	err := featureService.ValidateCallbackSecret(context.Background(), "shared-secret")
	if !errors.Is(err, ErrCommerceCallbackSecretNotConfigured) {
		t.Fatalf("ValidateCallbackSecret() error = %v, want %v", err, ErrCommerceCallbackSecretNotConfigured)
	}
}
