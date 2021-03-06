package boxversion_test

import (
  "testing"
  "naksu/boxversion"
)

func TestGetVagrantBoxType (t *testing.T) {
  tables := []struct {
    vagrantVersionString string
    vagrantBoxTypeString string
  }{
    {"", "-"},
    {"digabi/foobar", "-"},
    {"digabi/ktp-qa", "Abitti server"},
    {"digabi/ktp-k2018-45489", "Matric Exam server"},
  }

  for _, table := range tables {
    boxType := boxversion.GetVagrantBoxType(table.vagrantVersionString)
    if boxType != table.vagrantBoxTypeString {
      t.Errorf("GetVagrantBoxType gives '%s' instead of '%s'", boxType, table.vagrantBoxTypeString)
    }
  }
}

func TestGetVagrantBoxTypeIsAbitti (t *testing.T) {
  tables := []struct {
    vagrantVersionString string
    vagrantVersionIsAbitti bool
  }{
    {"", false},
    {"digabi/foobar", false},
    {"digabi/ktp-qa", true},
    {"digabi/ktp-k2018-45489", false},
  }

  for _, table := range tables {
    boxIsAbitti := boxversion.GetVagrantBoxTypeIsAbitti(table.vagrantVersionString)
    if boxIsAbitti != table.vagrantVersionIsAbitti {
      t.Errorf("GetVagrantBoxTypeIsAbitti fails with parameter '%s'", table.vagrantVersionString)
    }
  }
}

func TestGetVagrantBoxTypeIsMatriculationExam (t *testing.T) {
  tables := []struct {
    vagrantVersionString string
    vagrantVersionIsME bool
  }{
    {"", false},
    {"digabi/foobar", false},
    {"digabi/ktp-qa", false},
    {"digabi/ktp-k2018-45489", true},
  }

  for _, table := range tables {
    boxIsME := boxversion.GetVagrantBoxTypeIsMatriculationExam(table.vagrantVersionString)
    if boxIsME != table.vagrantVersionIsME {
      t.Errorf("GetVagrantBoxTypeIsMatriculationExam fails with parameter '%s'", table.vagrantVersionString)
    }
  }
}


func TestGetVagrantFileVersionAbitti (t *testing.T) {
  sampleAbittiVagrantFileContent := `
  # Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
  VAGRANTFILE_API_VERSION = "2"

  def calc_mem(system_mem_mb)
    mem_available_to_vm = (system_mem_mb * 0.74).floor
    lower_limit_mb = ((8192-1024)*0.74).floor # subtracting 1GB because of integrated graphics cards
    if mem_available_to_vm < lower_limit_mb
      $stderr.puts "Not enough memory left for virtual machine: #{mem_available_to_vm} MB. Required minimum memory for the computer is 8 GB"
      abort
    end
    mem_available_to_vm
  end

  def get_amount_of_cpus_and_system_ram
    # Give VM (host memory * 0.74) & all but 1 cpu (logical) cores
    host = RbConfig::CONFIG['host_os']

    cpus = nil
    mem = nil

    if host =~ /darwin/
      cpus = ` + "`" + `sysctl -n hw.logicalcpu_max` + "`" + `.to_i
      mem = ` + "`" + `sysctl -n hw.memsize` + "`" + `.to_i / 1024 / 1024
    elsif host =~ /linux/
      cpus = ` + "`" + `lscpu -p | awk -F',' '!/^#/{print $1}'| sort -u | wc -l` + "`" + `.to_i
      mem = ` + "`" + `grep 'MemTotal' /proc/meminfo | sed -e 's/MemTotal://' -e 's/ kB//'` + "`" + `.to_i / 1024
    elsif host =~ /mswin|mingw|cygwin/
      cpus = ` + "`" + `wmic cpu Get NumberOfLogicalProcessors` + "`" + `.split[1].to_i
      mem = ` + "`" + `wmic computersystem Get TotalPhysicalMemory` + "`" + `.split[1].to_i / 1024 / 1024
    end

    if cpus.nil?
      $stderr.puts "Could not determine the amount of cpus"
      abort
    end
    if mem.nil?
      $stderr.puts "Could not determine the amount of system memory"
      abort
    end
    [[cpus - 1, 2].max, calc_mem(mem)]
  end

  def get_nic_type
    if ENV['NIC']
      return ENV['NIC']
    else
      return 'virtio'
    end
  end

  cpus, mem = get_amount_of_cpus_and_system_ram()
  Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.boot_timeout = 300
    config.vm.box = "digabi/ktp-qa"
    config.vm.box_url = "https://s3-eu-west-1.amazonaws.com/static.abitti.fi/usbimg/qa/vagrant/metadata.json"
    config.vm.provider :virtualbox do |vb|
      vb.name = "SERVER7108X v57"
      vb.gui = true
      vb.customize ["modifyvm", :id, "--ioapic", "on"]
      vb.customize ["modifyvm", :id, "--cpus", cpus]
      vb.customize ["modifyvm", :id, "--memory", mem]
      vb.customize ["modifyvm", :id, "--nictype1", get_nic_type()]
      vb.customize ['modifyvm', :id, '--clipboard', 'bidirectional']
      vb.customize ["modifyvm", :id, "--vram", 24]
    end

    config.vm.synced_folder '~/ktp-jako', '/media/usb1', id: 'media_usb1'
    config.vm.synced_folder ".", "/vagrant", disabled: true
    config.vm.network "public_network", :adapter=>1, auto_config: false
  end
`

  sampleMebVagrantFileContent := `
  # Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
  VAGRANTFILE_API_VERSION = "2"

  def get_amount_of_cpus_and_system_ram
    # Give VM (host memory - 1.5 GB) & all but 1 cpu cores
    host = RbConfig::CONFIG['host_os']

    cpus = 3
    mem = 8192

    begin
      if host =~ /darwin/
        cpus = ` + "`" + `sysctl -n hw.physicalcpu_max` + "`" + `.to_i
        mem = ` + "`" + `sysctl -n hw.memsize` + "`" + `.to_i / 1024 / 1024
      elsif host =~ /linux/
        cpus = ` + "`" + `lscpu -p | awk -F',' '!/^#/{print $2}'| sort -u | wc -l` + "`" + `.to_i
        mem = ` + "`" + `grep 'MemTotal' /proc/meminfo | sed -e 's/MemTotal://' -e 's/ kB//'` + "`" + `.to_i / 1024
      elsif host =~ /mswin|mingw|cygwin/
        cpus = ` + "`" + `wmic cpu Get NumberOfCores` + "`" + `.split[1].to_i
        mem = ` + "`" + `wmic computersystem Get TotalPhysicalMemory` + "`" + `.split[1].to_i / 1024 / 1024
      end
    rescue
    end
    [[cpus - 1, 2].max, [mem - 1536, 6144].max]
  end

  cpus, mem = get_amount_of_cpus_and_system_ram()
  Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.boot_timeout = 300
    config.vm.box = "digabi/ktp-k2018-45489"
    config.vm.box_url = "https://s3-eu-west-1.amazonaws.com/static.abitti.fi/usbimg/k2018-45489/vagrant/metadata.json"
    config.vm.provider :virtualbox do |vb|
      vb.name = "SERVER7304M v37"
      vb.gui = true
      vb.customize ["modifyvm", :id, "--ioapic", "on"]
      vb.customize ["modifyvm", :id, "--cpus", cpus]
      vb.customize ["modifyvm", :id, "--memory", mem]
      vb.customize ["modifyvm", :id, "--nictype1", "virtio"]
      vb.customize ['modifyvm', :id, '--clipboard', 'bidirectional']
      vb.customize ["modifyvm", :id, "--vram", 24]
    end

    config.vm.synced_folder '~/ktp-jako', '/media/usb1', id: 'media_usb1'
    config.vm.synced_folder ".", "/vagrant", disabled: true
    config.vm.network "public_network", :adapter=>1, auto_config: false
  end
`

  tables := []struct {
    vagrantfileContent string
    versionString string
    versionType string
    humanReadableBoxType string
  }{
    {sampleAbittiVagrantFileContent, "SERVER7108X v57", "digabi/ktp-qa", "Abitti server"},
    {sampleMebVagrantFileContent, "SERVER7304M v37", "digabi/ktp-k2018-45489", "Matric Exam server"},
  }

  for _, table := range tables {
    versionType, versionString, _ := boxversion.GetVagrantVersionDetails(table.vagrantfileContent)
    humanReadableBoxType := boxversion.GetVagrantBoxType(versionType)

    if versionString != table.versionString {
      t.Errorf("GetVagrantVersionDetails returns wrong version string \"%s\" instead of \"%s\"", versionString, table.versionString)
    }

    if versionType != table.versionType {
      t.Errorf("GetVagrantVersionDetails returns wrong type string \"%s\" instead of \"%s\"", versionType, table.versionType)
    }

    if humanReadableBoxType != table.humanReadableBoxType {
      t.Errorf("GetVagrantBoxType return wrong result \"%s\" instead of \"%s\" when called with parameter \"%s\"", humanReadableBoxType, table.humanReadableBoxType, versionType)
    }
  }
}
