ovs-vsctl -- --id=@p get port fmdefaultvpc \
  -- --id=@m create mirror name=mirror-all \
     select-all=true \
     output-port=@p \
  -- set bridge nightlight mirrors=@m