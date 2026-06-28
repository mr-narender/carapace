function _carapace_identify_completer
  set --local data
  IFS='' set data (echo (commandline -cp)'' | sed "s/ \$/ ''/" | xargs example-multi $argv[1] _carapace fish 2>/dev/null)
  if [ $status -eq 1 ]
    IFS='' set data (echo (commandline -cp)"'" | sed "s/ \$/ ''/" | xargs example-multi $argv[1] _carapace fish 2>/dev/null)
    if [ $status -eq 1 ]
      IFS='' set data (echo (commandline -cp)'"' | sed "s/ \$/ ''/" | xargs example-multi $argv[1] _carapace fish 2>/dev/null)
    end
  end
  echo $data
end

complete -e "example-multi"
complete -c "example-multi" -f -a '(_carapace_identify_completer "example-multi")' -r
complete -e "identify"
complete -c "identify" -f -a '(_carapace_identify_completer "identify")' -r
complete -e "convert"
complete -c "convert" -f -a '(_carapace_identify_completer "convert")' -r

