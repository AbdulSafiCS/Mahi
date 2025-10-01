declare module "expo-secure-store" {
  export function isAvailableAsync(): Promise<boolean>;
  export function getItemAsync(
    key: string,
    options?: {
      keychainService?: string;
      requireAuthentication?: boolean;
      authenticationPrompt?: string;
    }
  ): Promise<string | null>;
  export function setItemAsync(
    key: string,
    value: string,
    options?: {
      keychainService?: string;
      requireAuthentication?: boolean;
      authenticationPrompt?: string;
    }
  ): Promise<void>;
  export function deleteItemAsync(
    key: string,
    options?: { keychainService?: string; requireAuthentication?: boolean }
  ): Promise<void>;
}
