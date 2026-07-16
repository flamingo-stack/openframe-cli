package argocd

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestManager_AdminPassword(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "argocd-initial-admin-secret", Namespace: "argocd"},
		Data:       map[string][]byte{"password": []byte("hunter2")},
	}
	m := &Manager{kubeClient: fake.NewSimpleClientset(secret)}

	pw, err := m.AdminPassword(context.Background())
	if err != nil {
		t.Fatalf("AdminPassword: %v", err)
	}
	if pw != "hunter2" {
		t.Fatalf("AdminPassword = %q, want %q", pw, "hunter2")
	}
}

func TestManager_AdminPassword_SecretMissing(t *testing.T) {
	m := &Manager{kubeClient: fake.NewSimpleClientset()}
	if _, err := m.AdminPassword(context.Background()); err == nil {
		t.Fatal("expected an error when the admin secret is absent")
	}
}

func TestManager_AdminPassword_NoPasswordField(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "argocd-initial-admin-secret", Namespace: "argocd"},
		Data:       map[string][]byte{"username": []byte("admin")},
	}
	m := &Manager{kubeClient: fake.NewSimpleClientset(secret)}
	if _, err := m.AdminPassword(context.Background()); err == nil {
		t.Fatal("expected an error when the password field is absent")
	}
}
