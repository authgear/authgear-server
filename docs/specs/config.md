# Configuration

  * [Configuration Conventions](#configuration-conventions)
    * [Prefer list over map](#prefer-list-over-map)
    * [Introduce flag only if necessary](#introduce-flag-only-if-necessary)
  * [References](#references)

## Configuration Conventions

This section outlines the configuration conventions Authgear must follow.

### Prefer list over map

Instead of

```yaml
login_id_keys:
  email:
    type: email
  phone:
    type: phone
  username:
    type: username
```

Do this

```yaml
login_id_keys:
- key: email
  type: email
- key: phone
  type: phone
- key: username
  type: username
```

### Introduce flag only if necessary

Add `enabled` or `disabled` flag only if necessary, such as toggling on/off of a feature.

## References

- [Kubernetes api conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
