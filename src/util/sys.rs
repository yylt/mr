use sysinfo::{self, System, Signal};
use rustix;

pub fn isroot() -> bool {
    match std::process::id() {
        0 => true,
        _ => rustix::process::geteuid().is_root(),
    }
}

pub fn send_hup_signal(pid: sysinfo::Pid) -> Option<bool> {
    match System::new_all().process(pid) {
        Some(process) =>  {
            process.kill_with(Signal::Hangup)
        }
        None => Some(false)
    }
    //rustix::process::kill_process(*pid, process::Signal::Hup);
}