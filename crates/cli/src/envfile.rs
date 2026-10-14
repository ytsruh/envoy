use std::{collections::BTreeMap, fs};

/// Parse an env file with standard dotenv semantics (via dotenvy):
/// single/double quotes, backslash escapes, `export` prefix,
/// inline ` #` comments, BOM, and multiline double-quoted values.
///
/// Notes on behavior worth knowing:
/// - Unquoted and double-quoted `$VAR`/`${VAR}` are interpolated against the
///   process environment, falling back to earlier keys in the same file and
///   then to empty (same as docker compose / python-dotenv). Single-quote a
///   value to keep `$` literal.
/// - Duplicate keys: last wins (standard dotenv behavior), with a stderr
///   warning naming the key and file.
/// - Empty values are rejected: the server requires non-empty values, so fail
///   fast naming the key instead of uploading and failing per key.
pub fn parse_env_file(path: &str) -> anyhow::Result<BTreeMap<String, String>> {
    let bytes = fs::read(path)
        .map_err(|e| anyhow::anyhow!("cannot read env file '{path}': {e}"))?;
    // dotenvy only strips the BOM on its load() path, not on the iterator
    // path, so strip it here to accept BOM-prefixed files explicitly.
    let bytes = bytes.strip_prefix(&[0xEF, 0xBB, 0xBF]).unwrap_or(&bytes);
    let mut map = BTreeMap::new();
    for item in dotenvy::from_read_iter(bytes) {
        let (k, v) = item
            .map_err(|e| anyhow::anyhow!("invalid format in env file '{path}': {e}"))?;
        if v.is_empty() {
            anyhow::bail!(
                "empty value for key '{k}' in env file '{path}': the server requires non-empty values"
            );
        }
        if map.contains_key(&k) {
            eprintln!("Warning: duplicate key '{k}' in {path}: using last value");
        }
        map.insert(k, v);
    }
    Ok(map)
}

/// Quote a value so standard dotenv parsers (docker compose, python-dotenv,
/// node dotenv, dotenvy) read back the exact same string:
/// - bare when it contains no special characters,
/// - single-quoted (literal everywhere) when possible,
/// - double-quoted with `\\`, `\"`, `\n`, `\$` escapes otherwise.
///
/// A literal carriage return is emitted as-is inside double quotes because
/// the `\r` escape is not understood by every parser.
fn quote_value(v: &str) -> String {
    if v.is_empty() {
        return String::new();
    }
    let needs_quote = v.chars().any(|c| {
        c == ' '
            || c == '\t'
            || c == '\n'
            || c == '\r'
            || c == '#'
            || c == '"'
            || c == '\''
            || c == '$'
            || c == '`'
            || c == '\\'
    });
    if !needs_quote {
        return v.to_string();
    }
    // Single quotes are literal in every major parser, but dotenvy treats a
    // backslash as escaping the closing quote when joining lines, so a value
    // ending in `\` (e.g. `trailing\`) would swallow the rest of the file.
    // Those values go through double quotes, where `\\` is exact.
    if !v.contains('\'') && !v.ends_with('\\') && !v.contains('\n') && !v.contains('\r') {
        return format!("'{v}'");
    }
    let mut out = String::with_capacity(v.len() + 2);
    out.push('"');
    for c in v.chars() {
        match c {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '$' => out.push_str("\\$"),
            _ => out.push(c),
        }
    }
    out.push('"');
    out
}

pub fn write_env_file(path: &str, vars: &BTreeMap<String, String>) -> anyhow::Result<()> {
    let mut out = String::new();
    for (k, v) in vars {
        out.push_str(&format!("{k}={}\n", quote_value(v)));
    }
    fs::write(path, out)?;
    Ok(())
}

pub fn sanitize_filename(name: &str) -> String {
    name.chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || c == '_' || c == '-' || c == '.' {
                c.to_ascii_lowercase()
            } else {
                '_'
            }
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    fn parse_content(content: &str) -> anyhow::Result<BTreeMap<String, String>> {
        let mut f = tempfile::NamedTempFile::new().expect("temp file");
        f.write_all(content.as_bytes()).expect("write fixture");
        f.flush().expect("flush fixture");
        parse_env_file(f.path().to_str().expect("temp path"))
    }

    fn parse_bytes(content: &[u8]) -> anyhow::Result<BTreeMap<String, String>> {
        let mut f = tempfile::NamedTempFile::new().expect("temp file");
        f.write_all(content).expect("write fixture");
        f.flush().expect("flush fixture");
        parse_env_file(f.path().to_str().expect("temp path"))
    }

    fn map(pairs: &[(&str, &str)]) -> BTreeMap<String, String> {
        pairs
            .iter()
            .map(|(k, v)| (k.to_string(), v.to_string()))
            .collect()
    }

    #[test]
    fn parses_basic_pairs_comments_and_blank_lines() {
        let got = parse_content(
            "# leading comment\n\nKEY=value\n\n  # indented comment\nOTHER=123\n",
        )
        .expect("parse");
        assert_eq!(got, map(&[("KEY", "value"), ("OTHER", "123")]));
    }

    #[test]
    fn parses_whitespace_around_key_and_equals() {
        let got = parse_content("  KEY  =  value  \n\tTAB\t=\ttabbed\t\n").expect("parse");
        assert_eq!(got, map(&[("KEY", "value"), ("TAB", "tabbed")]));
    }

    #[test]
    fn parses_single_double_and_empty_quotes() {
        let got = parse_content("A='single'\nB=\"double\"\nC='  padded  '\n").expect("parse");
        assert_eq!(
            got,
            map(&[("A", "single"), ("B", "double"), ("C", "  padded  ")])
        );
    }

    #[test]
    fn parses_export_prefix() {
        let got = parse_content("export KEY=value\nexport   SPACED  =  yes\n").expect("parse");
        assert_eq!(got, map(&[("KEY", "value"), ("SPACED", "yes")]));
    }

    #[test]
    fn parses_inline_comments() {
        // Space before # starts a comment; without the space it is literal.
        let got = parse_content("A=value # comment\nB=value#literal\n").expect("parse");
        assert_eq!(got, map(&[("A", "value"), ("B", "value#literal")]));
    }

    #[test]
    fn parses_escapes_in_double_quotes_and_bare_words() {
        let got =
            parse_content("A=\"line1\\nline2\"\nB=s\\ ix\nC=\"say \\\"hi\\\"\"\n").expect("parse");
        assert_eq!(
            got,
            map(&[
                ("A", "line1\nline2"),
                ("B", "s ix"),
                ("C", "say \"hi\""),
            ])
        );
    }

    #[test]
    fn parses_multiline_double_quoted_value() {
        let got = parse_content("A=\"line1\nline2\"\nB=after\n").expect("parse");
        assert_eq!(got, map(&[("A", "line1\nline2"), ("B", "after")]));
    }

    #[test]
    fn parses_equals_sign_inside_value() {
        let got = parse_content("CONN=host=db;port=5432\n").expect("parse");
        assert_eq!(got, map(&[("CONN", "host=db;port=5432")]));
    }

    #[test]
    fn parses_keys_with_dots_underscores_and_digits() {
        let got = parse_content("_A=1\nA1_B2=2\nfoo.bar=3\n").expect("parse");
        assert_eq!(got, map(&[("_A", "1"), ("A1_B2", "2"), ("foo.bar", "3")]));
    }

    #[test]
    fn parses_crlf_line_endings() {
        let got = parse_bytes(b"KEY=value\r\nOTHER=123\r\n").expect("parse");
        assert_eq!(got, map(&[("KEY", "value"), ("OTHER", "123")]));
    }

    #[test]
    fn parses_utf8_bom() {
        let mut bytes = vec![0xEF, 0xBB, 0xBF];
        bytes.extend_from_slice(b"KEY=value\n");
        let got = parse_bytes(&bytes).expect("parse");
        assert_eq!(got, map(&[("KEY", "value")]));
    }

    #[test]
    fn parses_unicode_values() {
        let got = parse_content("GREETING='héllo 🌍'\n").expect("parse");
        assert_eq!(got, map(&[("GREETING", "héllo 🌍")]));
    }

    #[test]
    fn single_quotes_keep_dollar_literal() {
        // No env dependence: single quotes block interpolation everywhere.
        let got = parse_content("A='s3cr$et'\nB='pre ${NOTHING} post'\n").expect("parse");
        assert_eq!(
            got,
            map(&[("A", "s3cr$et"), ("B", "pre ${NOTHING} post")])
        );
    }

    #[test]
    fn unquoted_dollar_interpolates_from_environment() {
        std::env::set_var("ENVOY_TEST_INTERP_XYZ", "42");
        let got = parse_content("B=pre-${ENVOY_TEST_INTERP_XYZ}-post\n").expect("parse");
        std::env::remove_var("ENVOY_TEST_INTERP_XYZ");
        assert_eq!(got, map(&[("B", "pre-42-post")]));
    }

    #[test]
    fn unquoted_dollar_interpolates_from_earlier_keys() {
        let got = parse_content("A=hello\nB=$A-world\n").expect("parse");
        assert_eq!(got, map(&[("A", "hello"), ("B", "hello-world")]));
    }

    #[test]
    fn duplicate_keys_last_wins() {
        let got = parse_content("KEY=first\nKEY=second\n").expect("parse");
        assert_eq!(got, map(&[("KEY", "second")]));
    }

    #[test]
    fn rejects_empty_values_naming_the_key() {
        for content in ["KEY=\n", "KEY=\"\"\n", "KEY='' \n", "KEY=  # comment\n"] {
            let err = parse_content(content).expect_err("must reject empty value");
            assert!(
                err.to_string().contains("KEY"),
                "error names the key: {err}"
            );
        }
    }

    #[test]
    fn rejects_malformed_lines() {
        // Each of these must fail loudly instead of uploading junk.
        // NOTE: `KEY="unterminated` spans lines, so keep each case in its
        // own file to avoid one case swallowing the rest of the fixture.
        // `KEY=value with spaces` errors because bare values cannot contain
        // spaces (quote them); `KEY=a\qb` has an invalid escape.
        let cases = [
            "=value\n",
            "123KEY=x\n",
            "NOEQUALS\n",
            "KEY value\n",
            "foo bar=x\n",
            "KEY=\"unterminated\n",
            "KEY='unterminated\n",
            "KEY=value with spaces\n",
            "KEY=a\\qb\n",
        ];
        for content in cases {
            parse_content(content)
                .expect_err(&format!("must reject malformed line: {content:?}"));
        }
    }

    #[test]
    fn rejects_missing_file_with_path() {
        let err = parse_env_file("/nonexistent-dir-xyz/missing.env").expect_err("must fail");
        assert!(
            err.to_string().contains("/nonexistent-dir-xyz/missing.env"),
            "error names the path: {err}"
        );
    }

    #[test]
    fn writer_leaves_safe_values_bare() {
        let dir = tempfile::tempdir().expect("tempdir");
        let path = dir.path().join("out.env");
        let vars = map(&[
            ("A", "simple"),
            ("B", "a=b"),
            ("C", "host:5432/db"),
            ("D", "under_score-1.0"),
        ]);
        write_env_file(path.to_str().unwrap(), &vars).expect("write");
        let raw = fs::read_to_string(&path).expect("read back");
        assert_eq!(raw, "A=simple\nB=a=b\nC=host:5432/db\nD=under_score-1.0\n");
    }

    #[test]
    fn writer_single_quotes_values_needing_protection() {
        let dir = tempfile::tempdir().expect("tempdir");
        let path = dir.path().join("out.env");
        let vars = map(&[
            ("A", "hello world"),
            ("B", "a#b"),
            ("C", "s3cr$t"),
            ("D", "say \"hi\""),
            ("E", "back\\slash"),
            ("F", "`cmd`"),
            ("G", "trailing "),
        ]);
        write_env_file(path.to_str().unwrap(), &vars).expect("write");
        let raw = fs::read_to_string(&path).expect("read back");
        assert_eq!(
            raw,
            "A='hello world'\nB='a#b'\nC='s3cr$t'\nD='say \"hi\"'\nE='back\\slash'\nF='`cmd`'\nG='trailing '\n"
        );
    }

    #[test]
    fn writer_double_quotes_when_single_quotes_or_newlines_present() {
        let dir = tempfile::tempdir().expect("tempdir");
        let path = dir.path().join("out.env");
        let vars = map(&[
            ("A", "it's"),
            ("B", "line1\nline2"),
            ("C", "both'and\"quotes"),
            ("D", "price $5\ntotal"),
            ("E", "trailing\\"),
        ]);
        write_env_file(path.to_str().unwrap(), &vars).expect("write");
        let raw = fs::read_to_string(&path).expect("read back");
        assert_eq!(
            raw,
            "A=\"it's\"\nB=\"line1\\nline2\"\nC=\"both'and\\\"quotes\"\nD=\"price \\$5\\ntotal\"\nE=\"trailing\\\\\"\n"
        );
    }

    #[test]
    fn write_then_parse_round_trips_tricky_values() {
        // The core compat property: anything pull writes, push reads back
        // byte-identical (empty values excluded: the server forbids them).
        let tricky = [
            "simple",
            "hello world",
            "  padded  ",
            "a#b",
            "a#",
            "#lead",
            "a=b=c",
            "tab\there",
            "s3cr$t",
            "${BRACED}",
            "`backtick`",
            "say \"hi\"",
            "it's",
            "both'and\"quotes",
            "back\\slash",
            "trailing\\",
            "line1\nline2",
            "quote\"and\nnewline",
            "dollar$and\nnewline",
            "héllo 🌍",
            "!bang",
            ";semi",
            "(paren)",
            "*star",
            "a; rm -rf /",
            "$(cmd)",
            "export",
            "with-hyphen_under.dot",
            "MiXeD123",
        ];
        let dir = tempfile::tempdir().expect("tempdir");
        let path = dir.path().join("roundtrip.env");
        let path_str = path.to_str().unwrap();
        let mut vars = BTreeMap::new();
        for (i, v) in tricky.iter().enumerate() {
            vars.insert(format!("KEY{i:02}"), v.to_string());
        }
        write_env_file(path_str, &vars).expect("write");
        let back = parse_env_file(path_str).expect("re-parse written file");
        assert_eq!(back, vars, "round trip must be lossless");
    }

    #[test]
    fn sanitize_filename_replaces_unsafe_chars() {
        assert_eq!(sanitize_filename("Prod-1"), "prod-1");
        assert_eq!(sanitize_filename("my env/prod"), "my_env_prod");
    }
}
