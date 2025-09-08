# Homebrew Formula for Octopus CLI
class OctopusCli < Formula
  desc "Dynamic Claude Code API management and forwarding tool"
  homepage "https://github.com/VibeAny/octopus-cli"
  url "https://github.com/VibeAny/octopus-cli/archive/v0.0.1.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256"
  license "MIT"
  head "https://github.com/VibeAny/octopus-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-X main.version=#{version}"), "./cmd"
    
    # Generate and install shell completions
    generate_completions_from_executable(bin/"octopus", "completion")
    
    # Install man page if available
    # man1.install "docs/octopus.1" if File.exist?("docs/octopus.1")
  end

  service do
    run [opt_bin/"octopus", "start", "--daemon"]
    keep_alive false
    log_path var/"log/octopus.log"
    error_log_path var/"log/octopus.log"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/octopus version")
    
    # Test configuration
    system bin/"octopus", "config", "add", "test", "https://example.com", "test-key"
    assert_predicate testpath/"configs/default.toml", :exist?
  end
end