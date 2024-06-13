use super::container;
use clap::Subcommand;

#[derive(Debug, Subcommand)]
#[command(about = "设置镜像")]
pub enum Set {
    #[command(name = "cnt")]
    #[command(alias = "container", about = "容器仓库相关镜像配置")]
    Cnt(container::Container),
}
