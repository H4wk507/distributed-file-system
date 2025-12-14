import { type ApiResponse } from "@/api/types";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useAxios } from "@/hooks/useAxios";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";

const registerSchema = z
  .object({
    username: z.string().min(3, "min. 3 znaki").max(20, "max. 20 znaków"),
    email: z.email("nieprawidłowy email"),
    password: z.string().min(8, "min. 8 znaków"),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "hasła nie pasują",
    path: ["confirmPassword"],
  });

type RegisterForm = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const navigate = useNavigate();

  const form = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      username: "",
      email: "",
      password: "",
      confirmPassword: "",
    },
  });

  const axios = useAxios();
  const registerUser = useMutation({
    mutationFn: async (data: {
      username: string;
      email: string;
      password: string;
    }) => {
      await axios.post<ApiResponse<void>>("/auth/register", data);
    },
    onSuccess: () => {
      navigate("/login");
    },
    onError: () => {
      form.setError("root", {
        message: "rejestracja nie powiodła się",
      });
    },
  });

  const onSubmit = async (data: RegisterForm) => {
    await registerUser.mutateAsync({
      username: data.username,
      email: data.email,
      password: data.password,
    });
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <div className="w-full max-w-xs">
        {/* Header */}
        <div className="mb-8 text-center">
          <div className="font-mono text-primary text-lg mb-2">[RSP]</div>
          <h1 className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
            utwórz konto
          </h1>
        </div>

        {/* Form */}
        <div className="border border-border bg-card">
          <div className="border-b border-border px-4 py-2">
            <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
              :: rejestracja ::
            </span>
          </div>

          <div className="p-4">
            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className="space-y-4"
              >
                {form.formState.errors.root && (
                  <div className="p-2 border border-destructive bg-destructive/10 text-destructive text-xs font-mono">
                    {form.formState.errors.root.message}
                  </div>
                )}

                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>nazwa użytkownika</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="jan_kowalski"
                          autoComplete="username"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>email</FormLabel>
                      <FormControl>
                        <Input
                          type="email"
                          placeholder="user@example.com"
                          autoComplete="email"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>hasło</FormLabel>
                      <FormControl>
                        <Input
                          type="password"
                          autoComplete="new-password"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="confirmPassword"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>potwierdź hasło</FormLabel>
                      <FormControl>
                        <Input
                          type="password"
                          autoComplete="new-password"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <Button
                  type="submit"
                  className="w-full"
                  disabled={registerUser.isPending}
                >
                  {registerUser.isPending ? (
                    <>
                      <Loader2 className="w-3.5 h-3.5 animate-spin" />
                      tworzenie...
                    </>
                  ) : (
                    "utwórz konto"
                  )}
                </Button>
              </form>
            </Form>
          </div>
        </div>

        <p className="mt-4 text-center text-xs text-muted-foreground font-mono">
          masz konto?{" "}
          <Link to="/login" className="text-primary hover:underline">
            zaloguj się
          </Link>
        </p>
      </div>
    </div>
  );
}
