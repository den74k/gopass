package leaf

import (
	"context"
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
)

// IsInitialized returns true if the store is properly initialized
func (s *Store) IsInitialized(ctx context.Context) bool {
	if s == nil || s.storage == nil {
		return false
	}
	return s.storage.Exists(ctx, s.idFile(ctx, ""))
}

// Init tries to initialize a new password store location matching the object
func (s *Store) Init(ctx context.Context, path string, ids ...string) error {
	if s.IsInitialized(ctx) {
		return fmt.Errorf(`found already initialized store at %q.
You can add secondary stores with gopass init --path <path to secondary store> --store <mount name>`, path)
	}

	// initialize recipient list
	recipients := make([]string, 0, len(ids))

	for _, id := range ids {
		if id == "" {
			continue
		}
		kl, err := s.crypto.FindRecipients(ctx, id)
		if err != nil {
			out.Errorf(ctx, "Failed to fetch public key for %q: %s", id, err)
			continue
		}
		if len(kl) < 1 {
			out.Errorf(ctx, "No useable keys for %q", id)
			continue
		}
		recipients = append(recipients, kl[0])
	}

	if len(recipients) < 1 {
		return fmt.Errorf("failed to initialize store: no valid recipients given in %+v", ids)
	}

	kl, err := s.crypto.FindIdentities(ctx, recipients...)
	if err != nil {
		return fmt.Errorf("failed to get available private keys: %w", err)
	}

	if len(kl) < 1 {
		return fmt.Errorf("none of the recipients has a secret key. You will not be able to decrypt the secrets you add")
	}

	if err := s.saveRecipients(ctx, recipients, "Initialized Store for "+strings.Join(recipients, ", ")); err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	return nil
}
