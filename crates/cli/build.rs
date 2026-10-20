fn main() {
    println!("cargo::rerun-if-env-changed=ENVOY_VERSION");
}
