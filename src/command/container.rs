use clap;
use lazy_static::lazy_static;
use std::collections::{HashMap,HashSet};
use std::error::Error;
use std::fs::{self,File,OpenOptions};
use std::os::unix::fs::PermissionsExt;
use sysinfo::{self, System};
use std::path::{Path};
use crate::util::sys;
use crate::merr;
use std::io::Write;
use serde_json::{self,Value};

lazy_static! {
    static ref REGISTRYS: HashMap<&'static str, &'static str> = {
        let mut m = HashMap::new();
        m.insert("docker", "docker.io");
        m.insert("gcr", "gcr.io");
        m.insert("ghcr", "ghcr.io");
        m.insert("k8s", "registry.k8s.io");
        m.insert("quay", "quay.io");
        m
    };
}

#[derive(Debug, clap::Args)]
pub struct Container {
    #[arg(short, long, value_name = "registry", default_value = "docker.io")]
    registry: Option<String>,

    #[arg(short, long, value_name = "mirror")]
    mirrors: Vec<String>,
}

impl Container {
    const MIRROR_KEY: &'static str = "registry-mirrors";
    const DEFAULT_REGISTRY: &'static str = "docker.io";
    pub fn update_mirror(&mut self) -> Result<(), Box<dyn Error>> {
        if !sys::isroot(){
            return Err(Box::new(merr!("以 root 或 sudo 运行")));
        }

        let mut dockerpid: Option<&sysinfo::Pid>= None; 
        let mut containerdpid: Option<&sysinfo::Pid>= None; 
        let sys = System::new_all();
        for (pid, process) in sys.processes() {
            // filter thread
            if process.thread_kind().is_some() {
                continue;
            }

            match process.name() {
                "dockerd" => {
                    dockerpid = Some(pid);
                }
                "containerd" => {
                    containerdpid =  Some(pid);
                }
                _ => (),
            }
        }

        if let Some(&pid) = dockerpid  {
            return self.docker(pid);
        }
        if let Some(&pid) = containerdpid  {
            return self.containerd(pid);
        }
       
        Err(Box::new(merr!("未找到运行时")))

    }

    fn docker(&self, pid: sysinfo::Pid) -> Result<(), Box<dyn Error>> {
        let reg = normalize_registry(&self.registry);
        if reg!=Container::DEFAULT_REGISTRY{
            return Err(Box::new(merr!("docker 只支持配置 docker.io")))
        }

        let path = Path::new("/etc/docker/daemon.json");
        fs::create_dir_all(path.parent().unwrap())?;

        let mut data = Value::Object(Default::default());
        let created = create_not_exist(path,0o755)?;
        
        if !created {
            data = serde_json::from_reader(File::open(path).unwrap())?;   
        }

        match data.get_mut(Container::MIRROR_KEY) {
            Some(mirror) => {
                if array_insert(mirror, &self.mirrors).is_some() {
                    println!("容器已配置镜像: {:?}, 跳过配置",self.mirrors);
                    return Ok(())
                }
            },
            None => {
                let mut mirror = serde_json::Map::new();
                
                mirror.insert(String::from(Container::MIRROR_KEY), Value::Array(self.mirrors.iter().map(|s| Value::String(normalize_mirror(s))).collect()));
                data.as_object_mut().unwrap().extend(mirror);
            },
        }
        let contents = serde_json::to_string_pretty(&data)?;
        let mut file = OpenOptions::new().write(true).open(path)?;
        file.write_all(contents.as_bytes())?;

        sys::send_hup_signal(pid);
        Ok(())
    }

    fn containerd(&self, _: sysinfo::Pid) -> Result<(), Box<dyn Error>> {
        Err(Box::new(merr!("未实现")))
    }
}


fn normalize_mirror(s: & str) -> String {
    
    if s.contains("//") {
        String::from(s)
    }else{
        format!("https://{s}")
    }
}

fn normalize_registry(s: &Option<String>) -> String {
    let default = String::from("docker.io");
    match s {
        Some(reg) => {
            REGISTRYS.get(reg.as_str()).map_or(default, |s| String::from(*s))
        },
        None => default
    }
}

// 检查文件并创建
// 存在时: 返回 false
// 不存在时: 返回 true或error
fn create_not_exist(f: &Path, mode: u32) -> Result<bool, std::io::Error> {
    
    match fs::metadata(f).is_err() {
        // 尝试创建
        true => match OpenOptions::new().create(true).truncate(true).open(f) {
            Ok(file) => {
                let mut perm = file.metadata().unwrap().permissions();
                perm.set_mode(mode);
                fs::set_permissions(f, perm).map(|_| false)
            },
            Err(e) => Err(e),
        }
        false =>  Ok(false)
    }
}


fn array_insert(val: &mut serde_json::Value, list:  &Vec<String> ) -> Option<()> {
    let mut ret = None;
    if let serde_json::Value::Array(ref mut arr) = *val {
        let mut unique_set: HashSet<String> = arr.iter().filter_map(|v| v.as_str().map(|s| s.to_string())).collect();
        for item in list {
            if unique_set.insert(item.clone()) {
                ret = Some(());
            }
        }
        ret?;

        arr.clear();
        for item in unique_set {
            arr.push(serde_json::Value::String(item));
        }
    }
    ret
}