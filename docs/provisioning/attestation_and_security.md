# Measured Boot Attestation

## Introduction

At the highest level, measured boot works by comparing (attesting) hash values (checksums) of certain critical components of a vulnerable system against some golden values.

In NICo ecosystem, this system has three components:

- Custom UEFI code on tenant machines, which performs measurements during machine boot and pushes those hash values (measurements) into so called PCR registers.
- `scout`, which runs inside discovery image, and which collects PCR values and submits them to `nico-api`.
- `nico-api`, which analyses PCR values received from `scout`, compares those values against golden values and makes a judgement call whether the measured boot is successful or not.

Measured Boot attestation happens during the ingestion of a managed host, and after a managed host has been released by a tenant.

## Measured Boot Scope and Its Effect

Measured Boot requires the presence of a TPM (Trusted Platform Module) chip on a managed host. DPUs do not support Measured Boot due to a lack of TPM hardware.

For the hosts the Measured Boot is configurable and can be turned off with `attestation_enabled` flag set to `false` in `carbide-api-site-config.toml`.

If Measured Boot is enabled and the attestation fails, the effect will be that `scout` will never get dispensed mTLS certificates that it needs to communicate with `nico-api`. Without mTLS certificates only gRPC calls not requiring mutual authentication will succeed, such as `DiscoverMachine` (which is where attestation happens) or `Version`.

## Measured Boot Operation

A prerequisite for any Measured Boot operation is the presence of a CA certificate. This certificate is needed to verify the authenticity of a TPM. When `scout` submits attestation info, it also includes its TPM's Endorsement Key (EK) certificate. `nico-api` verifies the authenticity of EK certificate (and thus of the TPM) by using a CA certificate.

When a machine is ingested into NICo, its DPU is being configured first (to provide network connectivity), and then its host is being provisioned second. The host boots into what is called a discovery, or a scout, image. Once booted into discovery image, the `scout` application runs and attempts to register the machine with the `nico-api` by making a gRPC `DiscoverMachine` call. In that call it supplies some attestation information. If the attestation is enabled, the `nico-api` replies with a challenge. The `scout` attempts to solve the challenge with the help of a TPM and, if successful, replies with a solved challenge and with the attestation information including PCR values. `nico-api` verifies the attestation information and the challenge, and, if both of those things are successful, it dispenses mTLS certificates to `scout`, which are necessary for all future communication. If the attestation is unsuccessful, the `scout` will keep calling `DiscoverMachine` for the next ten years at a rate of one call per minute.

Besides the ingestion path, the Measured Boot will also occur when an assigned machine is released back into the pool, or when attestation status for the machine changes while the machine is in the Ready state.

## Configuring Measured Boot

Once `attestation_enabled` flag is set, the `nico-api` will attempt to do attestation for all hosts.

By default, if none of the approval strategies have been enacted as discussed below in the section [Configuring nico-api for Attestation](#configuring-nico-api-for-attestation), the machines will be stuck in `Measuring/WaitingForMeasurements` state.

## Installing CA Certificates

First, a CA certificate needs to be installed. This is done using `nico-admin-cli`'s `tpm-ca` commands.

- `tpm-ca show-unmatched-ek` will show all EK certificates/machines for which there is no matching CA certificate (a machine will be in `Failed/MeasurementsCAValidationFailed` state in this instance). A link to download the CA certificate will typically be shown also. It is important that whoever is adding CA certificates makes sure they are downloaded from a reputable source - the need for a human verification is the reason why this operation was not automated in the first place.
- `tpm-ca show` will display all installed CA certificates.
- `tpm-ca add --filename <ca_cert>` will add a given CA certificate to `nico-api`.
- `tpm-ca delete --ca-id <ca_cert_id>` will delete a given CA certificate from `nico-api`. The `ca_cert_id` can be retrieved from the output of `tpm-ca show` command.
- `tpm-ca add-bulk --dirname <dir_path>` will add all certificates in a given directory to `nico-api`.

CA certificates must have extension `.pem`, `.cer` or `.der` and must be in corresponding format. A validation is done in `nico-admin-cli` to confirm that this is the case.

## Configuring nico-api for Attestation

The attestation on the `nico-api` side is done with the help of the following entities, which are all present in `nico-api` DB and can be accessed with `nico-admin-cli`:

- `report` - this is what `scout` sends to `nico-api`, including PCR values. It is accessible under `measured-boot report` subcommand.
- `profile` - this is combination of properties that describe a type of a machine. It is used to group identical machines together. It is accessible under `measured-boot profile` subcommand.
- `bundle` - this is a set of golden values against which attestation is being made. Each bundle is linked to a profile. It is accessible under `measured-boot bundle` subcommand.
- `journal` - this is an "umbrella" entity linking report to a possible bundle and to a profile. It is accessible under `measured-boot report` subcommand and will contain the measurement status of a machine.

Once a report is sent for a given machine, it is saved in a DB, and `nico-api` attempts to find a corresponding machine profile, failing that it will create a new profile. After the report is saved and profile is found or created, it will attempt to find the matching bundle and save a journal entry linking all of the above. If a bundle has been identified, the machine's attestation status will be `Measured` and the process returns.

If a bundle has not been identified, the `nico-api` will attempt to create a bundle from the report in one of two ways:

1. It will check if a given machine (or any machine in a site) has been marked for auto-approval.
2. Failing that, it will check if a given profile has been marked for auto-approval.

If a bundle could not be created in any of the two ways above, the attestation will initially fail and `scout` will keep making `DiscoverMachine` calls every minute for the next ten years.

One way to resolve this situation is to mark a given machine for one shot auto approval with:

```sh
measured-boot site trusted-machine approve <machine id> oneshot --pcr-registers <list of PCR registers>
```

In this way a given machine will be marked for automatic approval one time only. This will be enough to create an initial bundle. An alternative way would be to promote the existing report with:

```sh
measured-boot report promote <REPORT ID>
```

Similarly, a given profile can be marked for one shot auto approval with:

```sh
measured-boot site trusted-profile approve <profile id> oneshot --pcr-registers <list of PCR registers>
```

In both of the above cases, the `--pcr-registers` argument is optional. If omitted, all available PCR registers will be selected.

If a site is to be left running in permissive mode, an auto approve can be configured for all machines on a site with:

```sh
measured-boot site trusted-machine approve * persist
```

> Note: the `*` has a tendency to confuse bash scripts, if one is to script this command.

Similarly, a `persist` auto approval can be set for a profile.

In order to make it easier set up measured boot on new sites, it is possible to export an existing site with `measured-boot site export` command and then import the values with `measured-boot site import` command on a different site.

For the full breakdown of possible options, please browse `nico-admin-cli measured-boot` subcommand, where you will discover some other options not discussed here, such as e.g. being able to manually set the state of a bundle.

## NICo Web UI
In the NICo Web UI, there is a Section named "Attestations", where the attestations for all machines are listed. It makes it easier to see the attestations' status at a glance, drill down into individual reports and create bundles out of reports.

## Practical Application
One issue with Measured Boot is that the PCR values can change quite frequently. The operator needs to establish which PCR values are necessary for a particular monitoring use case, and which ones of those are stable enough to include in the bundle.

## Troubleshooting

When a site had just been set up and the attestation had been enabled, all machines will fail attestation unless it had been configured. In order to enable the "accept anything" type of attestation, the easiest way is to execute a command similar to this one:

```sh
nico-admin-cli -c https://nico-api.forge/ measured-boot site trusted-machine approve \* persist --pcr-registers="0,3,5,6"
```

This will insert a rule to automatically approve all reports for all machines on site `az32`, that rule will persist and bundles will be created out of reports using PCR registers 0, 3, 5 and 6. Obviously, such attestation is not enforcing any specific PCR register values, so it should be used with caution.

If a machine is stuck in `Measuring/WaitingForMeasurements` state, this means the measurements have not been sent by `scout`. This may be due to the problem booting into discovery/scout image, or maybe due to some problem with the `scout` itself.

Other Measured Boot error states are:

- `PendingBundle` - this means attestation is being enforced and measurements from a given host do not have matching golden values. Please refer to above sections for details on creating a new bundle from the latest report. Also, if a closest match has been identified, it will be shown in the Attestation table in NICo UI. Using `nico-admin-cli measured-boot bundle find-closest-match` will achieve the same result.
- `Failed/MeasurementsCAValidationFailed` indicates that a corresponding CA certificate has not been found for a EK cert that is returned by this machine's TPM. Please refer to above sections on details about installing missing CA certificates.
