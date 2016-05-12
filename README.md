# gosyncd

A simple configuration file synchronization tool.

_gosyncd_ is a daemon and a command-line utility to run on multiple servers where configuration files should be identical. It will analyze possible discrepancies and replicate them on all the hosts in the group.

It can be compared to csync2 with the main differences being:

 * TLS mutual authentication between all the nodes in a group
 * No state database used to track changes

## Example configuration

The program configuration is written with Hashicorp's HCL format:

```
bind_address = ":8080"

# List all the nodes in the group, self IP address will be blacklisted
remotes = [
  "172.17.0.2:8080",
  "172.17.0.3:8080"
]

tls_ca_certificate = "/etc/gosyncd/gosyncd.crt"
tls_certificate = "/etc/gosyncd/gosyncd.crt"
tls_private_key = "/etc/gosyncd/gosyncd.key"

directory "/etc/haproxy" {
  exclude = [
    "/etc/haproxy/errors"
  ]
}

directory "/foo/bar" {
}
```

### Notes on TLS certificates

_gosyncd_ uses TLS to encrypt communications and for mutual authentication (client **and** server). That means the certificate must be valid for the hostnames (or IP addresses) of the host.

You can generate a single key pair for all the hosts in a group, using DNS or IP Subject Alternative Names.

You can also use self-signed certificates by setting the *tls_ca_certificate* to the same TLS certificate.

## Example usage

```bash
# On all the nodes in the group
$ gosyncd -config /etc/gosyncd/config.hcl

# On the node where files were edited, to see what would be done
$ gosyncd -config /etc/gosyncd/config.hcl sync -dry
INFO[0000] synchronization dry run                      
INFO[0000] 2016/12/05 12:14:27 +0000: [M] 172.17.0.4:8080 /etc/haproxy/haproxy.cfg updated
INFO[0000] 2016/12/05 12:14:27 +0000: [-] 172.17.0.4:8080 /etc/haproxy/test deleted
INFO[0000] 2016/12/05 12:14:27 +0000: [+] 172.17.0.4:8080 /foo/bar/coucou added

# On the node where files were edited, to sync edits
$ gosyncd -config /etc/gosyncd/config.hcl sync
INFO[0000] 2016/12/05 12:14:27 +0000: [M] 172.17.0.4:8080 /etc/haproxy/haproxy.cfg updated
INFO[0000] 2016/12/05 12:14:27 +0000: [-] 172.17.0.4:8080 /etc/haproxy/test deleted
INFO[0000] 2016/12/05 12:14:27 +0000: [+] 172.17.0.4:8080 /foo/bar/coucou added
```

## Notes

 * Performance could still be enhanced: for now, one request is made per pending update, a future release will merge them into a single one
 * The daemon does not handle creating backup copies of updated file (**yet**)
 * The daemon does not handle running actions on updates (**yet**)
