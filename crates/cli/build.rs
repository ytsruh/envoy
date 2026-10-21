fn main() {
    // `src/main.rs` reads the release version via `option_env!("ENVOY_VERSION")`
    // (stamped from the git tag by the release workflow). Cargo does not track
    // arbitrary env vars in its fingerprints, so without this a rebuild with a
    // different ENVOY_VERSION would be skipped as "fresh" and `envoy --version`
    // would report a stale value. This opts the variable into change tracking.
    println!("cargo::rerun-if-env-changed=ENVOY_VERSION");
}
