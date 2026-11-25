pub fn temporary_directory() -> tempfile::TempDir {
    tempfile::Builder::new()
        .prefix("agent-print-session-test-")
        .tempdir()
        .unwrap()
}
