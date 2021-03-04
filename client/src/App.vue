<script>
import { h, computed } from "vue"
import routes from "./routes"

export default {
  setup() {
    const currentRoute = computed(() => window.location.pathname)
    const currentComponent = computed(() => routes[currentRoute.value] || "")

    return () => h(currentComponent.value)
  }
}

</script>

