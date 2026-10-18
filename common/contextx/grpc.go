package contextx

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func ExtractFromGRPCMetadata(ctx context.Context) *Builder {
	builder := NewBuilder()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return builder
	}
	bindTrace(builder, func(key string) string {
		values := md.Get(key)
		if len(values) > 0 {
			return values[0]
		}
		return ""
	}, true)
	return builder
}

func InjectIntoGRPCMetadata(ctx context.Context) context.Context {
	md := metadata.MD{}
	injectTrace(FromContext(ctx), func(key, value string) {
		if value != "" {
			md.Set(key, value)
		}
	}, true)

	if existingMD, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(md, existingMD)
	}
	if cred, ok := CredentialsFromContext(ctx); ok && cred != nil {
		if cred.APIKey != "" {
			md.Set(MetadataExchangeAPIKey, cred.APIKey)
		}
		if cred.SecretKey != "" {
			md.Set(MetadataExchangeAPISecret, cred.SecretKey)
		}
		if cred.Passphrase != "" {
			md.Set(MetadataExchangePassphrase, cred.Passphrase)
		}
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func CredentialsFromIncomingMetadata(ctx context.Context) (*Credentials, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, false
	}
	key := firstMD(md, MetadataExchangeAPIKey)
	secret := firstMD(md, MetadataExchangeAPISecret)
	pass := firstMD(md, MetadataExchangePassphrase)
	if key == "" && secret == "" {
		return nil, false
	}
	return &Credentials{APIKey: key, SecretKey: secret, Passphrase: pass}, true
}

func firstMD(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
