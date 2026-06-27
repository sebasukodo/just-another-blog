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
