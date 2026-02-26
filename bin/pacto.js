#!/usr/bin/env node

const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const https = require("node:https");

const pkg = require("../package.json");

const PLATFORM = {
  linux: "linux",
  darwin: "darwin",
};

const ARCH = {
  x64: "amd64",
  arm64: "arm64",
};

function fail(msg) {
  console.error(`pacto wrapper error: ${msg}`);
  process.exit(1);
}

function download(url, outPath) {
  return new Promise((resolve, reject) => {
    const req = https.get(url, (res) => {
      if (
        res.statusCode &&
        res.statusCode >= 300 &&
        res.statusCode < 400 &&
        res.headers.location
      ) {
        res.resume();
        download(res.headers.location, outPath).then(resolve).catch(reject);
        return;
      }

      if (res.statusCode !== 200) {
        reject(new Error(`download failed ${res.statusCode}: ${url}`));
        res.resume();
        return;
      }

      const file = fs.createWriteStream(outPath);
      res.pipe(file);
      file.on("finish", () => {
        file.close(() => resolve());
      });
      file.on("error", reject);
    });
    req.on("error", reject);
  });
}

function sha256File(filePath) {
  const hash = crypto.createHash("sha256");
  hash.update(fs.readFileSync(filePath));
  return hash.digest("hex");
}

function extractExpectedChecksum(checksumsPath, artifactName) {
  const content = fs.readFileSync(checksumsPath, "utf8");
  for (const line of content.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const parts = trimmed.split(/\s+/);
    if (parts.length >= 2 && parts[1] === artifactName) {
      return parts[0].toLowerCase();
    }
  }
  return "";
}

function run(cmd, args) {
  const r = spawnSync(cmd, args, { stdio: "inherit" });
  if (r.error) throw r.error;
  if (typeof r.status === "number" && r.status !== 0) {
    process.exit(r.status);
  }
  if (r.signal) {
    process.kill(process.pid, r.signal);
  }
}

function cacheRoot() {
  if (process.env.PACTO_CACHE_DIR) return process.env.PACTO_CACHE_DIR;
  if (process.env.XDG_CACHE_HOME) {
    return path.join(process.env.XDG_CACHE_HOME, "pacto");
  }
  return path.join(os.homedir(), ".cache", "pacto");
}

async function ensureBinary() {
  const platform = PLATFORM[process.platform];
  const arch = ARCH[process.arch];
  if (!platform || !arch) {
    fail(`unsupported platform ${process.platform}/${process.arch}`);
  }

  const repo = process.env.PACTO_REPO || "triple0-labs/pacto-spec";
  const versionRaw = process.env.PACTO_VERSION || pkg.version;
  const version = versionRaw.startsWith("v") ? versionRaw.slice(1) : versionRaw;
  const tag = `v${version}`;
  const artifact = `pacto_${version}_${platform}_${arch}.tar.gz`;
  const baseURL = `https://github.com/${repo}/releases/download/${tag}`;

  const installDir = path.join(cacheRoot(), version, `${platform}_${arch}`);
  const binaryPath = path.join(installDir, "pacto");
  if (fs.existsSync(binaryPath)) {
    fs.chmodSync(binaryPath, 0o755);
    return binaryPath;
  }

  fs.mkdirSync(installDir, { recursive: true });
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "pacto-npm-"));
  const artifactPath = path.join(tmpDir, artifact);
  const checksumsPath = path.join(tmpDir, "checksums.txt");

  try {
    await download(`${baseURL}/${artifact}`, artifactPath);
    await download(`${baseURL}/checksums.txt`, checksumsPath);

    const expected = extractExpectedChecksum(checksumsPath, artifact);
    if (!expected) {
      fail(`checksum entry not found for ${artifact}`);
    }
    const actual = sha256File(artifactPath);
    if (actual !== expected) {
      fail(`checksum mismatch for ${artifact}`);
    }

    run("tar", ["-xzf", artifactPath, "-C", installDir]);
    if (!fs.existsSync(binaryPath)) {
      fail(`binary not found after extract: ${binaryPath}`);
    }
    fs.chmodSync(binaryPath, 0o755);
    return binaryPath;
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

async function main() {
  const binary = await ensureBinary();
  const args = process.argv.slice(2);
  const child = spawnSync(binary, args, { stdio: "inherit" });
  if (child.error) {
    fail(child.error.message);
  }
  process.exit(child.status === null ? 1 : child.status);
}

main().catch((err) => fail(err.message));
