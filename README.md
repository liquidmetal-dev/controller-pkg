## controller-pkg

This repo contains packages which are common between various liquid metal components.

So far we have:
- `client`: a client to `flintlock`
- `services/microvm`: a service to create microvms via the `flintlock` client
- `types/microvm`: Microvm object types

They are all **separate** modules, thus a change in one required in another
would require you to commit the first before you update the other.
For example: if you make a change in `types/microvm` which is relevant in `services/microvm`,
you would need to merge that change first at which point you could update the module
file for `services/microvm`.

This was done so that dependants will only pull in the modules they require.

(We will need to think about versioning the `types`, at least, at some point.)
