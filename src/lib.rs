pub mod command;
pub mod util;

#[macro_export]
macro_rules! merr {
    ($msg:expr) => {
        $crate::util::error::MyError {
            message: $msg.to_string(),
        }
    };
}
