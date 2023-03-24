package constants

// PreNetworkConfigScript script runs on hosts before network manager service starts in order to apply
// user's provided network configuration on the host.
// If the user provides static network configuration, the network config files will be stored in directory
// /etc/assisted/network in the following structure:
// /etc/assisted/network/
//                    +-- host1
//                    |      +--- *.nmconnection
//                    |      +--- mac_interface.ini
//                    +-- host2
//                          +--- *.nmconnection
//                          +--- mac_interface.ini
// 1. *.nmconnections - files generated by nmstate based on yaml files provided by the user
// 2. mac_interface.ini - the file contains mapping of mac-address to logical interface name.
//    There are two usages for the file:
//    1. Map logical interface name to MAC Address of the host. The logical interface name is a
//       name provided by the user for the interface. It will be replaced by the script with the
//       actual network interface name.
//    2. Identify the host directory which belongs to the current host by matching a MAC Address
//       from the mapping file with host network interfaces.
//
// Applying the network configuration of each host will be done by:
// 1. Associate the current host with its matching hostX directory. The association will be done by
//    matching host's mac addresses with those in mac_interface.ini.
// 2. Replace logical interface name in nmconnection files with the interface name as set on the host
// 3. Rename nmconnection files to start with the interface name (instead of the logical interface name)
// 4. Copy the nmconnection files to /NetworkManager/system-connections/
const PreNetworkConfigScript = `#!/bin/bash

# The directory that contains nmconnection files of all nodes
NMCONNECTIONS_DIR=/etc/assisted/network
MAC_NIC_MAPPING_FILE=mac_interface.ini

if [[ ! -d "$NMCONNECTIONS_DIR" ]]
then
  echo "Error (exiting): Expected to find the directory $NMCONNECTIONS_DIR on the host but this was not present."
  exit 0
fi

# A map of host mac addresses to interface names
declare -A host_macs_to_hw_iface

# The directory that contains nmconnection files for the current host
host_dir=""

# The mapping file of the current host
mapping_file=""

# A nic-to-mac map created from the mapping file associated with the host
declare -A logical_nic_mac_map

# Find destination directory based on ISO mode
if [[ -f /etc/initrd-release ]]; then
  ETC_NETWORK_MANAGER="/run/NetworkManager/system-connections"
else
  ETC_NETWORK_MANAGER="/etc/NetworkManager/system-connections"
fi
echo "Info: ETC_NETWORK_MANAGER was set to $ETC_NETWORK_MANAGER"

# Create a map of host mac addresses to their network interfaces
function map_host_macs_to_interfaces() {
  SYS_CLASS_NET_DIR='/sys/class/net'
  for nic in $( ls $SYS_CLASS_NET_DIR )
  do
    mac=$(cat $SYS_CLASS_NET_DIR/$nic/address | tr '[:lower:]' '[:upper:]')
    host_macs_to_hw_iface[$mac]=$nic
  done
}

function find_host_directory_by_mac_address() {
  for d in $(ls -d ${NMCONNECTIONS_DIR}/host*)
  do

    mapping_file="${d}/${MAC_NIC_MAPPING_FILE}"
    if [[ ! -f "$mapping_file" ]]
    then
      echo "Warning: Mapping file $mapping_file is missing. Skipping on directory $d"
      continue
    fi

    if [[ -z "$(ls -A $d/*.nmconnection)" ]]
    then
      echo "Warning: Host directory does not contain any nmconnection files, skipping"
      continue
    fi

    # check if mapping file contains mac-address that exists on the current host
    for mac_address in $(cat $mapping_file | cut -d= -f1 | tr '[:lower:]' '[:upper:]')
    do
      if [[ ! -z "${host_macs_to_hw_iface[${mac_address}]:-}" ]]
      then
        host_dir=$(mktemp -d)
        cp ${d}/* $host_dir
        echo "Info: Found host directory: $d, configuration copied to $host_dir"
        return
      fi
      echo "Info: No match found in host macs for '$mac_address'"
    done
  done

  if [ -z "$host_dir" ]
  then
    echo "Error: None of the host directories contained a mac-address to host mapping for the current host"
    // We should not exit the script here in any fashion as the Dracut initqueue handler will not run any subsequent scripts if there is an exit.
  fi
}

function set_logical_nic_mac_mapping() {
  echo "Info: Checking '$mapping_file' for logical nic to mac mappings"
  # initialize logical_nic_mac_map with mapping file entries
  readarray -t lines < "${mapping_file}"
  for line in "${lines[@]}"
  do
    mac=${line%%=*}
    nic=${line#*=}
    logical_nic_mac_map[$nic]=${mac^^}
  done
}

# Replace logical interface name in nmconnection files with the interface name from the mapping file
# of host's directory. Replacement is done based on mac-address matching
function update_interface_names_by_mapping_file() {
  echo "Info: Updaing logical interface names in nmconnection files with interface name from mapping files"

  # iterate over host's nmconnection files and replace logical interface name with host's nic name
  for nmconn_file in $(ls -1 ${host_dir}/*.nmconnection)
  do
    echo "Info: Updaing logical interface name in nmconnection files with interface name from mapping file $host_dir/$nmconn_file"
    # iterate over mapping to find nmconnection files with logical interface name
    for nic in "${!logical_nic_mac_map[@]}"
    do
      mac=${logical_nic_mac_map[$nic]}

      # the pattern should match '=eth0' (interface name) or '=eth0.' (for vlan devices)
      if grep -q -e "=$nic$" -e "=$nic\." "$nmconn_file"
      then
        # get host interface name
        host_iface=${host_macs_to_hw_iface[$mac]}
        if [ -z "$host_iface" ]
        then
          echo "Warning: Mapping file contains MAC Address '$mac' (for logical interface name '$nic') that doesn't exist on the host"
          continue
        fi

        # replace logical interface name with host interface name
        sed -i -e "s/=$nic$/=$host_iface/g" -e "s/=$nic\./=$host_iface\./g" $nmconn_file
        echo "Info: Using logical interface name '$nic' for interface with Mac address '$mac', updated $nmconn_file)"
      fi
    done
  done
}

function copy_nmconnection_files_to_nm_config_dir() {
  echo "Info: Copying nmconnection files to $ETC_NETWORK_MANAGER"
  for nmconn_file in $(ls -1 ${host_dir}/*.nmconnection)
  do
    # rename nmconnection files based on the actual interface name
    filename=$(basename $nmconn_file)
    prefix="${filename%%.*}"
    extension="${filename#*.}"
    if [ ! -z "${logical_nic_mac_map[$prefix]}" ]
    then
      dir_path=$(dirname $nmconn_file)
      mac_address=${logical_nic_mac_map[$prefix]}
      host_iface=${host_macs_to_hw_iface[$mac_address]}
      if [ ! -z "$host_iface" ]
      then
        mv $nmconn_file "${dir_path}/${host_iface}.${extension}"
        echo "Info: Copied $nmconn_file to $dir_path/$host_iface/$extension"
        continue
      fi
      echo "Warning: Mapping for '$mac_address' was not present while attempting to copy $nmconn_file to $dir_path/$host_iface/$extension"
    fi
  done

  cp ${host_dir}/*.nmconnection ${ETC_NETWORK_MANAGER}/
  echo "Info: Copied all nmconnection files from $host_dir to $ETC_NETWORK_MANAGER"
}

echo "PreNetworkConfig Start"
map_host_macs_to_interfaces
find_host_directory_by_mac_address

// Make sure we do not run any of the following functions if there was no matching config.
if [ "$host_dir" ]
  then
    # remove default connection file create by NM(nm-initrd-generator). This is a WA until
    # NM is back to supporting priority between nmconnections
    rm -f ${ETC_NETWORK_MANAGER}/*
    echo "Removing default connection files in '$ETC_NETWORK_MANAGER'"
    set_logical_nic_mac_mapping
    update_interface_names_by_mapping_file
    copy_nmconnection_files_to_nm_config_dir
fi
echo "PreNetworkConfig End"
`
