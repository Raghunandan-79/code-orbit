interface OrbitIconProps {
  size?: number;
  className?: string;
}

export function OrbitIcon({ size = 32, className }: OrbitIconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      aria-hidden="true"
    >
      {/* Outer orbit ring */}
      <ellipse
        cx="16"
        cy="16"
        rx="13"
        ry="6"
        stroke="#6d6aff"
        strokeWidth="1.5"
        strokeDasharray="4 2"
        opacity="0.5"
        transform="rotate(-20 16 16)"
      />
      {/* Inner orbit ring */}
      <ellipse
        cx="16"
        cy="16"
        rx="8"
        ry="3.5"
        stroke="#a5a3ff"
        strokeWidth="1.25"
        opacity="0.7"
        transform="rotate(60 16 16)"
      />
      {/* Core star */}
      <circle cx="16" cy="16" r="3" fill="#6d6aff" />
      <circle cx="16" cy="16" r="1.5" fill="white" opacity="0.8" />
      {/* Orbiting dot */}
      <circle cx="27" cy="13" r="2" fill="#a5a3ff" opacity="0.9" />
    </svg>
  );
}
