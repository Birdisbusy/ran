import semver from 'semver';

import { VMBackend, BackendEvents } from './backend';
import { ServiceEntry } from './client';

import EventEmitter from '@/utils/eventEmitter';

export {
  BackendSettings, FailureDetails, RestartReasons, State,
  BackendError as KubernetesError, BackendProgress as KubernetesProgress,
} from './backend';
export type { ServiceEntry } from './client';

/**
 * VersionEntry describes a version of K3s.
 */
export interface VersionEntry {
  /** The version being described. This includes any build-specific data. */
  version: semver.SemVer;
  /** A string describing the channels that include this version, if any. */
  channels?: string[];
}

/**
 * KubernetesBackendEvents describes the events that may be emitted by a
 * Kubernetes backend (as an EventEmitter).  Each property name is the name of
 * an event, and the property type is the type of the callback function expected
 * for the given event.
 */
export interface KubernetesBackendEvents extends BackendEvents {
  /**
   * Emitted when the set of Kubernetes services has changed.
   */
  'service-changed'(services: ServiceEntry[]): void;

  /**
   * Emitted when an error related to the port forwarding server has occurred.
   */
  'service-error'(service: ServiceEntry, errorMessage: string): void;

  /**
   * Emitted when the versions of Kubernetes available has changed.
   */
  'versions-updated'(): void;

  /**
   * Emitted when k8s is running on a new port
   */
  'current-port-changed'(port: number): void;

  /**
   * Emitted when the checkForExistingKimBuilder setting pref changes
   */
  'kim-builder-uninstalled'(): void;
}

export interface KubernetesBackend extends VMBackend, EventEmitter<KubernetesBackendEvents>, KubernetesBackendPortForwarder {
  /**
   * The versions that are available to install, sorted as would be displayed to
   * the user.
   */
  availableVersions: Promise<VersionEntry[]>;

  /**
   * Used to let the UI know whether it was sent all potentially supported k8s versions.
   * If this returns true, it means we're only telling the UI which versions we have cached.
   */
  cachedVersionsOnly(): Promise<boolean>;

  /** The version of Kubernetes that is currently installed. */
  version: string;

  /**
   * The port the Kubernetes server will listen on; this may not reflect the
   * port correctly if the server is not active.
   */
  readonly desiredPort: number;

  /**
   * Fetch the list of services currently known to Kubernetes.
   * @param namespace The namespace containing services; omit this to
   *                  return services across all namespaces.
   */
  listServices(namespace?: string): ServiceEntry[];
}

export interface KubernetesBackendPortForwarder {
  /**
   * Forward a single service port, returning the resulting local port number.
   * @param namespace The namespace containing the service to forward.
   * @param service The name of the service to forward.
   * @param k8sPort The internal port of the service to forward.
   * @param hostPort The host port to listen on for the forwarded port. Pass 0 for a random port.
   * @returns The port listening on localhost that forwards to the service.
   */
  forwardPort(namespace: string, service: string, k8sPort: number | string, hostPort: number): Promise<number | undefined>;

  /**
   * Cancel an existing port forwarding.
   * @param namespace The namespace containing the service to forward.
   * @param service The name of the service to forward.
   * @param k8sPort The internal port of the service to forward.
   * @param hostPort The host port to listen on for the forwarded port. Pass 0 for a random port.
   */
  cancelForward(namespace: string, service: string, k8sPort: number | string): Promise<void>;
}
