import CountriesTimezonesData from './countries_timezones.json'
import map from 'lodash/map'

// convert to arrays
type Country = {
  name: string
  abbr: string
  zones: string[]
}

export const Timezones = map(CountriesTimezonesData.zones, (x) => x)
export const TimezonesFormOptions = map(Timezones, (x) => {
  return {
    value: x,
    label: x
  }
})
export const CountriesMap: Record<string, Country> = CountriesTimezonesData.countries
export const Countries = map(CountriesTimezonesData.countries, (x) => x)
export const CountriesFormOptions = map(CountriesTimezonesData.countries, (x) => {
  return {
    value: x.abbr,
    label: x.abbr + ' - ' + x.name
  }
})

export default CountriesTimezonesData
