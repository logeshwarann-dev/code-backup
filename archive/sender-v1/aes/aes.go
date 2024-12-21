package aes

/*
#cgo CFLAGS: -I"../openssl/include"
#cgo LDFLAGS: -L"../openssl" -l:libcrypto.so.3
#include <openssl/evp.h>
#include <string.h>
#include <stdlib.h>

typedef struct CipherContext {
    EVP_CIPHER_CTX *encrypt_ctx;
    EVP_CIPHER_CTX *decrypt_ctx;
    unsigned char key[32];
    unsigned char iv[16];
} CipherContext;

// Initialize the encryption and decryption contexts
int aes_gcm_init(CipherContext *ctx, const unsigned char *key, const unsigned char *iv) {
    ctx->encrypt_ctx = EVP_CIPHER_CTX_new();
    ctx->decrypt_ctx = EVP_CIPHER_CTX_new();

    if (!ctx->encrypt_ctx || !ctx->decrypt_ctx) return -1;

    memcpy(ctx->key, key, 32);
    memcpy(ctx->iv, iv, 16);

    // Initialize encryption and decryption operations
    if (EVP_EncryptInit_ex(ctx->encrypt_ctx, EVP_aes_256_gcm(), NULL, NULL, NULL) != 1) return -1;
    if (EVP_CIPHER_CTX_ctrl(ctx->encrypt_ctx, EVP_CTRL_GCM_SET_IVLEN, 16, NULL) != 1) return -1;
    if (EVP_EncryptInit_ex(ctx->encrypt_ctx, NULL, NULL, ctx->key, ctx->iv) != 1) return -1;

    if (EVP_DecryptInit_ex(ctx->decrypt_ctx, EVP_aes_256_gcm(), NULL, NULL, NULL) != 1) return -1;
    if (EVP_CIPHER_CTX_ctrl(ctx->decrypt_ctx, EVP_CTRL_GCM_SET_IVLEN, 16, NULL) != 1) return -1;
    if (EVP_DecryptInit_ex(ctx->decrypt_ctx, NULL, NULL, ctx->key, ctx->iv) != 1) return -1;

    return 0;
}

unsigned char *aes_gcm_encrypt_update(CipherContext *ctx, const unsigned char *plaintext, int plaintext_len, int *ciphertext_len) {
    unsigned char *ciphertext = (unsigned char *)malloc(plaintext_len);
    int len;

    if (EVP_EncryptUpdate(ctx->encrypt_ctx, ciphertext, &len, plaintext, plaintext_len) != 1) {
        free(ciphertext);
        return NULL;
    }

    *ciphertext_len = len;
    return ciphertext;  // Return dynamically allocated ciphertext
}

unsigned char *aes_gcm_decrypt_update(CipherContext *ctx, const unsigned char *ciphertext, int ciphertext_len, int *plaintext_len) {
    unsigned char *plaintext = (unsigned char *)malloc(ciphertext_len);
    int len;

    if (EVP_DecryptUpdate(ctx->decrypt_ctx, plaintext, &len, ciphertext, ciphertext_len) != 1) {
        free(plaintext);
        return NULL;
    }

    *plaintext_len = len;
    return plaintext;  // Return dynamically allocated plaintext
}

// Clean up function to release the encryption and decryption contexts
void aes_gcm_cleanup(CipherContext *ctx) {
    EVP_CIPHER_CTX_free(ctx->encrypt_ctx);
    EVP_CIPHER_CTX_free(ctx->decrypt_ctx);
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

// CipherContext is a Go wrapper for the C CipherContext struct
type CipherContext struct {
	ctx *C.CipherContext
}

// NewCipherContext initializes a new CipherContext
func NewCipherContext(key, iv []byte) (*CipherContext, error) {
	cCtx := C.malloc(C.size_t(unsafe.Sizeof(C.CipherContext{})))
	if cCtx == nil {
		return nil, errors.New("failed to allocate memory for CipherContext")
	}
	ctx := (*C.CipherContext)(cCtx)

	if C.aes_gcm_init(ctx, (*C.uchar)(unsafe.Pointer(&key[0])), (*C.uchar)(unsafe.Pointer(&iv[0]))) != 0 {
		C.free(cCtx)
		return nil, errors.New("failed to initialize CipherContext")
	}

	return &CipherContext{ctx: ctx}, nil
}

// EncryptUpdate performs encryption
func (c *CipherContext) EncryptUpdate(plaintext []byte) ([]byte, error) {
	var ciphertextLen C.int
	ciphertext := C.aes_gcm_encrypt_update(c.ctx, (*C.uchar)(unsafe.Pointer(&plaintext[0])), C.int(len(plaintext)), &ciphertextLen)
	if ciphertext == nil {
		return nil, errors.New("encryption failed")
	}

	// Convert C pointer to Go byte slice
	goCiphertext := C.GoBytes(unsafe.Pointer(ciphertext), ciphertextLen)

	defer C.free(unsafe.Pointer(ciphertext)) // Free ciphertext after use
	return goCiphertext, nil
}

// DecryptUpdate performs decryption
func (c *CipherContext) DecryptUpdate(ciphertext []byte) ([]byte, error) {
	var plaintextLen C.int
	plaintext := C.aes_gcm_decrypt_update(c.ctx, (*C.uchar)(unsafe.Pointer(&ciphertext[0])), C.int(len(ciphertext)), &plaintextLen)
	if plaintext == nil {
		return nil, errors.New("decryption failed")
	}

	// Convert C pointer to Go byte slice
	goPlaintext := C.GoBytes(unsafe.Pointer(plaintext), plaintextLen)

	defer C.free(unsafe.Pointer(plaintext)) // Free plaintext after use
	return goPlaintext, nil
}

// Cleanup releases resources held by the CipherContext
func (c *CipherContext) Cleanup() {
	C.aes_gcm_cleanup(c.ctx)
	C.free(unsafe.Pointer(c.ctx)) // Free CipherContext memory
}
