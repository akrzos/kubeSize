# kubeSize

kubeSize is a kubernetes CLI plugin to easily expose sizing and capacity data for Kubernetes clusters.

# Develop

## Compile

```
$ make bin
```

## Run

```
$ ./bin/capacity
```

## Install as a plugin

```
cp bin/capacity /usr/local/bin/kubectl-capacity
```

## Use

```
$ kubectl capacity
Capacity exposes size and capacity data for Kubernetes clusters

Usage:
  capacity [command]

Available Commands:
  cluster     Get cluster size and capacity
  help        Help about any command
  namespace   Get namespace size
  node        Get individual node capacity
  node-role   Get cluster capacity grouped by node role

Flags:
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default HTTP cache directory (default "/Users/akrzos/.kube/http-cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
  -h, --help                           help for capacity
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
  -r, --readable                       Human readable resources (CPU in Cores, Memory in GiB)
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use

Use "capacity [command] --help" for more information about a command.
```
