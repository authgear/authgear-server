(* Skygear module for Augeas
 Adapted from puppet.aug
*)
module Skygear =
  autoload xfm

(************************************************************************
 * INI File settings
 *************************************************************************)
let comment    = IniFile.comment IniFile.comment_re IniFile.comment_default
let sep        = IniFile.sep "=" "="


(************************************************************************
 *                        ENTRY
 *************************************************************************)
let entry   = IniFile.indented_entry IniFile.entry_re sep comment


(************************************************************************
 *                        RECORD
 *************************************************************************)
let title   = IniFile.indented_title IniFile.record_re
let record  = IniFile.record title entry


(************************************************************************
 *                        LENS & FILTER
 *************************************************************************)
let lns     = IniFile.lns record comment

let filter = (incl "/home/ubuntu/myapp/development.ini")

let xfm = transform lns filter
