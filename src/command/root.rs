use super::set;
use clap::Parser;

#[derive(Debug, Parser)]
#[command(name = "mr")]
#[command(bin_name = "mr")]
#[command(about = "mirror tool for repo, language, software(docker, container), etc...", long_about = None)]
pub enum Root {
    #[command(subcommand)]
    Set(set::Set),
}
