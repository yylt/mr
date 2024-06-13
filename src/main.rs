use clap::Parser;
use mr::command::{root, set};

fn main() {
    match root::Root::parse() {
        root::Root::Set(x) => match x {
            set::Set::Cnt(mut cnt) => {
               cnt.update_mirror().map_or_else(|err| println!("{err}")
               ,|_| { println!("更新成功")} );
            }
        }, // _ => print!("no match"),
    }
}
