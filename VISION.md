# Yar Vision

## The Promise

Yar is the version-controlled, executable contract between a team's local development environment and its production platform.

Declare a system once. Yar makes the right fleet available for the environment you choose: on a developer laptop, in a development cluster, or as production-ready deployment artifacts.

Everyone's local environment is the same in intent, and that intent is production.

## Why Yar Exists

Modern service development accumulates local rituals instead of a shared system model. One developer has a Compose wrapper. Another has a shell script, a hand-maintained `.env`, and a few undocumented host entries. A third knows how to start the VPN, seed Redis, and connect the right network. Those arrangements can work for one person, but they do not form a reliable team platform.

The failure appears when the platform changes. A teammate adds a dependency or secret. Someone returns to a project after weeks away. A frontend, backend, and several sidecars need to work together. The result is familiar: missing `.env` values passed in direct messages, stale setup notes, brittle local networking, and a system that works only on the machine of the person who last touched it.

Production has an explicit model: services, dependencies, identities, configuration, secrets, and target infrastructure. Local development deserves the same clarity.

## The Model

`yar.yaml` is the project's declarative system model. It describes:

- Services and supporting sidecars
- Dependency order and replica intent
- Stable service identities and network reachability
- Non-secret configuration and environment-specific targets
- Secret references, never secret values
- The packs that turn service intent into runnable infrastructure

Packs make the model portable. A pack describes a service once and can produce local Docker behavior, Helm charts, or Kubernetes manifests. The local runtime and the production output are different realizations of the same declared intent, not independently maintained approximations.

For example, if production services communicate through `redis.foo`, local development should use that same service identity. The implementation may differ behind the scenes, but application configuration should not need to learn a second topology just because it is running on a laptop.

## The Experience

Yar should feel like a calm fleet control plane, not a pile of infrastructure tools exposed to every developer.

```sh
yar project init
yar project edit
yar hoist
yar dock
yar scuttle
```

`yar hoist` is deliberately simple at the surface. Beneath it, Yar validates the system model, checks the complete secret inventory, prepares the local substrate, establishes declared service reachability, starts dependencies in order, and reports the fleet's state.

The developer should not need to remember which shell script starts this project, whether it needs Redis or RabbitMQ, which Docker network to join, how hostnames resolve, or which new secret a teammate added last week. Yar should tell them what is missing, why it matters, and how to fix it.

## The Secret Contract

Secrets are a team contract, not a file-sharing workflow.

A service declares the secret references it requires. When a developer pulls that change and runs `yar hoist`, Yar discovers every required reference before starting anything. If a secret is missing, Yar stops with an actionable inventory and a path to resolve it through the configured provider.

Secret values are not committed to the repository, copied through chat, or normalized into shared `.env` files. They are resolved at runtime from approved encrypted stores and injected through the correct mechanism for the target platform.

This turns a late runtime surprise into an immediate, understandable part of changing the system model.

## What Yar Replaces

Yar replaces the coordination burden of bespoke local-environment tooling:

- Project-specific Compose wrappers and cleanup scripts
- Machine-specific setup rituals
- Ad-hoc Ansible used to bridge gaps between developer machines
- Shared `.env` files and direct-message secret distribution
- Separate, drifting local and production definitions
- Undocumented networking workarounds for reaching local sidecars

It preserves the useful outcomes of those tools: local sidecars, production-shaped names, host reachability, lifecycle operations, and safe cleanup. It makes those outcomes declarative, inspectable, testable, portable, and owned by the project.

## How Yar Works With the Platform

Yar does not try to replace Docker, Kubernetes, Helm, Colima, VPN software, or GitOps.

It works with them through native SDKs where possible and narrow, structured adapters where an SDK does not exist. Docker runs local workloads. Kubernetes and Helm receive generated production artifacts. Colima and networking infrastructure provide the local substrate. GitOps remains responsible for production deployment.

Yar owns the project's declared resources, not the whole machine. It must identify what it created, clean up only what it owns, and leave unrelated Docker projects, networks, host entries, and platform infrastructure untouched.

## The Value

For developers, Yar replaces setup archaeology with a reliable command path.

For teams, it turns local development into a shared, versioned contract instead of a collection of personal conventions.

For platform engineers, it creates a single model that can generate consistent local and production-facing artifacts.

For new team members, it makes a complete environment available from the repository instead of from oral history.

## Decision Test

Every Yar feature should strengthen this promise:

> Does this make the declared system more faithfully portable between team members and environments, without exposing secrets or taking ownership of resources Yar did not create?

If the answer is no, it is not central to Yar.
