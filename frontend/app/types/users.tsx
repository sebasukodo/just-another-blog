export interface User {
  username: string;
  email: string;
  token: string;
  bio: string | null;
  image: string | null;
}

export interface UserResponse {
  user: User;
}

export interface UserUpdate {
  username: string;
  email: string;
  bio: string | null;
  image: string | null;
}
